package domain

import "context"

type Role struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type Permission struct {
	ID   int    `db:"id"`
	Name string `db:"name"`
}

type RolePermissions struct {
	Role
	Permissions []Permission
}

type UserRoles struct {
	User
	Roles []string
}

type RBACCacheRepository interface {
	Get(ctx context.Context, roles []string) ([]string, error)
	Set(ctx context.Context, rolePermissions []RolePermissions) error
}

type RBACService interface {
	CheckPermission(ctx context.Context, roles []string, permissionName string) (bool, error)
	LoadRolePermissions(ctx context.Context) error
}

type RBACRepository interface {
	// GetRoles(ctx context.Context) ([]Role, error)
	// CreateRole(ctx context.Context, roleName string) (Role, error)
	// UpdateRole(ctx context.Context, role Role) (Role, error)
	// DeleteRole(ctx context.Context, id int) error
	// GetPermissions(ctx context.Context) ([]Permission, error)
	// CreatePermission(ctx context.Context, permissionName string) (Permission, error)
	// UpdatePermission(ctx context.Context, permission Permission) (Permission, error)
	// DeletePermission(ctx context.Context, id int) error
	GetRolePermissions(ctx context.Context) ([]RolePermissions, error)
	// GetRolePermissionsByRoleID(ctx context.Context, roleID int) ([]RolePermissions, error)
	// CreateRolePermissions(ctx context.Context, roleName string, permissions []Permission) (RolePermissions, error)
	// UpdateRolePermissions(ctx context.Context, roleName string, permissions []Permission) (RolePermissions, error)
}
