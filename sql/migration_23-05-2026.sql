CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(150) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Mengunci keunikan email di tingkat profil utama
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique ON users (LOWER(email));



CREATE TABLE IF NOT EXISTS user_authentications (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    provider VARCHAR(20) NOT NULL,      -- 'local' atau 'google'
    provider_key VARCHAR(255) NOT NULL,  -- Jika 'local' diisi email, jika 'google' diisi Google UID
    password_hash VARCHAR(255) NULL,    -- Hanya diisi jika provider = 'local'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Relasi ke tabel users
    CONSTRAINT fk_auth_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
        
    -- Aturan Bisnis 1: User tidak boleh punya lebih dari satu metode untuk provider yang sama
    CONSTRAINT uq_user_provider UNIQUE(user_id, provider),
    
    -- Aturan Bisnis 2: Satu ID Google / satu email lokal tidak boleh dipakai oleh 2 akun yang berbeda
    CONSTRAINT uq_provider_key UNIQUE(provider, provider_key)
);


CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    token VARCHAR(255) NOT NULL,
    device_info VARCHAR(255) NULL,       -- Menyimpan browser/nama device user agent
    is_revoked BOOLEAN DEFAULT FALSE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- Relasi ke tabel users
    CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) 
        REFERENCES users(id) ON DELETE CASCADE,
        
    -- Token harus acak dan unik secara global
    CONSTRAINT uq_refresh_token_string UNIQUE(token)
);


-- Mempercepat query pencarian token saat frontend melakukan request POST /auth/refresh
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_lookup ON user_refresh_tokens (token) WHERE is_revoked = FALSE;

-- Mempercepat proses pencarian relasi session per-user jika user ingin melihat daftar device aktif
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON user_refresh_tokens (user_id);


-- PostgreSQL membutuhkan fungsi khusus untuk memperbarui kolom updated_at secara otomatis
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