package repositories

import (
	"auth/domain"
	"context"
	"sync"
)

type inMemoryRBACRepository struct {
	rolePermissions []domain.RolePermissions
	sync.RWMutex
}

func NewInMemoryRBACRepository() domain.RBACCacheRepository {
	return &inMemoryRBACRepository{}
}

func (r *inMemoryRBACRepository) Get(ctx context.Context, roleName []string) ([]string, error) {
	r.RLock()
	defer r.RUnlock()

	var permissions []string
	for _, rp := range r.rolePermissions {
		for _, role := range roleName {
			if rp.Role.Name == role {
				for _, perm := range rp.Permissions {
					permissions = append(permissions, perm.Name)
				}
			}
		}
	}
	return permissions, nil
}

func (r *inMemoryRBACRepository) Set(ctx context.Context, rolePermissions []domain.RolePermissions) error {
	r.Lock()
	defer r.Unlock()

	r.rolePermissions = rolePermissions

	return nil
}
