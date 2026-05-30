package repositories

import (
	"auth/domain"
	"context"

	"github.com/doug-martin/goqu"
)

type rbacRepository struct {
	db *goqu.Database
}

// GetRolePermissions implements [domain.RBACRepository].
func (r *rbacRepository) GetRolePermissions(ctx context.Context) ([]domain.RolePermissions, error) {
	type rolePermission struct {
		RoleID         int    `db:"role_id"`
		RoleName       string `db:"role_name"`
		PermissionID   int    `db:"permission_id"`
		PermissionName string `db:"permission_name"`
	}
	var rolePermissions []rolePermission

	err := r.db.From(goqu.I("role_permissions").As("rp")).
		LeftJoin(
			goqu.I("roles").As("r"),
			goqu.On(goqu.Ex{"rp.role_id": goqu.I("r.id")}),
		).
		LeftJoin(
			goqu.I("permissions").As("p"),
			goqu.On(goqu.Ex{"rp.permission_id": goqu.I("p.id")}),
		).
		Select(
			goqu.I("r.id").As("role_id"),
			goqu.I("r.name").As("role_name"),
			goqu.I("p.id").As("permission_id"),
			goqu.I("p.name").As("permission_name"),
		).
		ScanStructsContext(ctx, &rolePermissions)

	if err != nil {
		return nil, err
	}

	rolePermissionsMap := make(map[int]domain.RolePermissions)
	for _, rp := range rolePermissions {
		if _, exists := rolePermissionsMap[rp.RoleID]; !exists {
			rolePermissionsMap[rp.RoleID] = domain.RolePermissions{
				Role: domain.Role{
					ID:   rp.RoleID,
					Name: rp.RoleName,
				},
				Permissions: []domain.Permission{},
			}
		}
		rolePerms := rolePermissionsMap[rp.RoleID]

		if rp.PermissionID != 0 {
			rolePerms.Permissions = append(
				rolePerms.Permissions, domain.Permission{
					ID:   rp.PermissionID,
					Name: rp.PermissionName,
				})
		}
		rolePermissionsMap[rp.RoleID] = rolePerms
	}

	var result []domain.RolePermissions
	for _, rolePerm := range rolePermissionsMap {
		result = append(result, rolePerm)
	}

	return result, nil
}

func NewRBACRepository(db *goqu.Database) domain.RBACRepository {
	return &rbacRepository{
		db: db,
	}
}
