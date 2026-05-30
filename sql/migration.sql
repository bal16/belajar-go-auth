CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Locking email uniqueness at the main profile level
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (LOWER(email));



CREATE TABLE IF NOT EXISTS user_authentications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    provider VARCHAR(20) NOT NULL,      -- 'local' or 'google'
    provider_key VARCHAR(255) NOT NULL,  -- If 'local' filled with email, if 'google' filled with Google UID
    password_hash VARCHAR(255) NULL,    -- Only filled if provider = 'local'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- relation to users
    CONSTRAINT fk_auth_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
        
    -- Bussiness Logic 1: User cannot have more than one method for the same provider
    CONSTRAINT uq_user_provider UNIQUE(user_id, provider),
    
    -- Business Logic 2: One Google ID TOKEN / one local email cannot be used by 2 different accounts
    CONSTRAINT uq_provider_key UNIQUE(provider, provider_key)
);


CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    --implement cuid2
    id VARCHAR(255) PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    -- device_info VARCHAR(255) NULL,       -- Store the user agent browser/device name
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- relation to users
    CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
        
    -- Token must be unique and globally random
    CONSTRAINT uq_refresh_token_string UNIQUE(token)
);


-- Speeding up token lookup queries when the frontend makes a POST /auth/refresh request
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_lookup ON user_refresh_tokens (id) WHERE is_revoked = FALSE;

-- Speeding up the process of searching for session relationships per-user if the user wants to view a list of active devices
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON user_refresh_tokens (user_id);


-- PostgreSQL requires a special function to automatically update the updated_at column
CREATE OR REPLACE FUNCTION update_modified_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_user_modtime
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_modified_column();

CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INT NOT NULL,
    role_id INT NOT NULL,
    PRIMARY KEY (user_id, role_id),
    
    CONSTRAINT fk_user_role_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_user_role_role FOREIGN KEY (role_id) 
        REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INT NOT NULL,
    permission_id INT NOT NULL,
    PRIMARY KEY (role_id, permission_id),
    
    CONSTRAINT fk_role_perm_role FOREIGN KEY (role_id) 
        REFERENCES roles(id) ON DELETE CASCADE,
        
    CONSTRAINT fk_role_perm_permission FOREIGN KEY (permission_id) 
        REFERENCES permissions(id) ON DELETE CASCADE
);

-- Seed initial roles and permissions
INSERT INTO "roles" (name) VALUES ('super_admin'), ('admin'), ('user');

INSERT INTO "permissions" (name) VALUES 
    ('user.create'),
    ('user.read'),
    ('user.update'),
    ('user.delete'),
    ('role.create'),
    ('role.read'),
    ('role.update'),
    ('role.delete'),
    ('permission.create'),
    ('permission.read'),
    ('permission.update'),
    ('permission.delete');




INSERT INTO "role_permissions" (role_id, permission_id)
-- 1. All permission for Super Admin
SELECT 
    (SELECT id FROM "roles" WHERE name = 'super_admin'), id 
FROM "permissions"

UNION ALL

-- 2. Permission read & update for users & roles resource for Admin
SELECT 
    (SELECT id FROM "roles" WHERE name = 'admin'), id
FROM "permissions"
-- WHERE action IN ('read', 'update') AND resource IN ('users', 'roles')
WHERE name IN ('user.read', 'user.update', 'role.read', 'role.update', 'permission.read', 'permission.update')

UNION ALL

-- 3. Permission read for users resource for regular User
SELECT 
    (SELECT id FROM "roles" WHERE name = 'user'), id
FROM "permissions" 
-- WHERE action = 'read' AND resource = 'users'
WHERE name IN ('user.read', 'permission.read', 'role.read')

ON CONFLICT DO NOTHING;