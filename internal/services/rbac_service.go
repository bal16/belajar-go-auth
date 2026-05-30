package services

import (
	"auth/domain"
	"context"
	"errors"
	"slices"
)

type rbacService struct {
	rbacCacheRepository domain.RBACCacheRepository
	rbacRepo            domain.RBACRepository
}

func NewRBACService(
	rbacCacheRepository domain.RBACCacheRepository,
	rbacRepo domain.RBACRepository,
) domain.RBACService {
	s := &rbacService{
		rbacCacheRepository: rbacCacheRepository,
		rbacRepo:            rbacRepo,
	}

	s.LoadRolePermissions(context.Background())
	return s
}

func (s *rbacService) CheckPermission(ctx context.Context, roles []string, permissionName string) (bool, error) {
	//get user roles from http memory -- in controller/handler/middleware
	//get roles permissions from cache
	permissions, err := s.rbacCacheRepository.Get(ctx, roles)
	if err != nil {
		return false, err
	}
	//check if user has permission
	if slices.Contains(permissions, permissionName) {
		return true, nil
	}

	return false, nil
}

func (s *rbacService) LoadRolePermissions(ctx context.Context) error {
	rolePermissions, err := s.rbacRepo.GetRolePermissions(ctx)
	if err != nil {
		return err
	}
	err = s.rbacCacheRepository.Set(ctx, rolePermissions)
	if err != nil {
		return errors.New("failed to set role permissions in cache")
	}
	return nil
}
