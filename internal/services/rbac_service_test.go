package services_test

import (
	"context"
	"errors"
	"testing"

	"auth/domain"
	"auth/internal/services"
)

// --- Mock RBACRepository ---

type mockRBACRepository struct {
	getRolePermissionsFunc func(ctx context.Context) ([]domain.RolePermissions, error)
}

func (m *mockRBACRepository) GetRolePermissions(ctx context.Context) ([]domain.RolePermissions, error) {
	if m.getRolePermissionsFunc != nil {
		return m.getRolePermissionsFunc(ctx)
	}
	return nil, nil
}

// --- Mock RBACCacheRepository ---

type mockRBACCacheRepository struct {
	getFunc func(ctx context.Context, roleName []string) ([]string, error)
	setFunc func(ctx context.Context, rolePermissions []domain.RolePermissions) error
}

func (m *mockRBACCacheRepository) Get(ctx context.Context, roleName []string) ([]string, error) {
	if m.getFunc != nil {
		return m.getFunc(ctx, roleName)
	}
	return nil, nil
}

func (m *mockRBACCacheRepository) Set(ctx context.Context, rolePermissions []domain.RolePermissions) error {
	if m.setFunc != nil {
		return m.setFunc(ctx, rolePermissions)
	}
	return nil
}

// --- Tests ---

func TestNewRBACService(t *testing.T) {
	mockRepo := &mockRBACRepository{
		getRolePermissionsFunc: func(ctx context.Context) ([]domain.RolePermissions, error) {
			return []domain.RolePermissions{}, nil
		},
	}
	mockCache := &mockRBACCacheRepository{
		setFunc: func(ctx context.Context, rolePermissions []domain.RolePermissions) error {
			return nil
		},
	}

	// Ensure initialization works and doesn't panic
	svc := services.NewRBACService(mockCache, mockRepo)
	if svc == nil {
		t.Error("Expected RBACService to be initialized, got nil")
	}
}

func TestRBACService_LoadRolePermissions(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockRBACRepository{
			getRolePermissionsFunc: func(ctx context.Context) ([]domain.RolePermissions, error) {
				return []domain.RolePermissions{
					{Role: domain.Role{Name: "admin"}},
				}, nil
			},
		}
		mockCache := &mockRBACCacheRepository{
			setFunc: func(ctx context.Context, rolePermissions []domain.RolePermissions) error {
				if len(rolePermissions) != 1 {
					t.Errorf("Expected 1 role permission to be set, got %d", len(rolePermissions))
				}
				return nil
			},
		}
		svc := services.NewRBACService(mockCache, mockRepo)

		err := svc.LoadRolePermissions(context.Background())
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("GetRolePermissions Failed", func(t *testing.T) {
		mockRepo := &mockRBACRepository{
			getRolePermissionsFunc: func(ctx context.Context) ([]domain.RolePermissions, error) {
				return nil, errors.New("db query error")
			},
		}
		mockCache := &mockRBACCacheRepository{}

		svc := services.NewRBACService(mockCache, mockRepo)
		err := svc.LoadRolePermissions(context.Background())
		if err == nil || err.Error() != "db query error" {
			t.Errorf("Expected 'db query error', got %v", err)
		}
	})

	t.Run("Cache Set Failed", func(t *testing.T) {
		mockRepo := &mockRBACRepository{
			getRolePermissionsFunc: func(ctx context.Context) ([]domain.RolePermissions, error) {
				return []domain.RolePermissions{}, nil
			},
		}
		mockCache := &mockRBACCacheRepository{
			setFunc: func(ctx context.Context, rolePermissions []domain.RolePermissions) error {
				return errors.New("redis connection lost")
			},
		}
		svc := services.NewRBACService(mockCache, mockRepo)

		err := svc.LoadRolePermissions(context.Background())
		if err == nil || err.Error() != "failed to set role permissions in cache" {
			t.Errorf("Expected 'failed to set role permissions in cache', got %v", err)
		}
	})
}

func TestRBACService_CheckPermission(t *testing.T) {
	mockRepo := &mockRBACRepository{}

	t.Run("Has Permission", func(t *testing.T) {
		mockCache := &mockRBACCacheRepository{
			getFunc: func(ctx context.Context, roleName []string) ([]string, error) {
				return []string{"user.read", "user.write", "role.read"}, nil
			},
		}
		svc := services.NewRBACService(mockCache, mockRepo)

		hasPerm, err := svc.CheckPermission(context.Background(), []string{"admin"}, "user.read")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if !hasPerm {
			t.Error("Expected to have permission, got false")
		}
	})

	t.Run("No Permission", func(t *testing.T) {
		mockCache := &mockRBACCacheRepository{
			getFunc: func(ctx context.Context, roleName []string) ([]string, error) {
				return []string{"user.read"}, nil
			},
		}
		svc := services.NewRBACService(mockCache, mockRepo)

		hasPerm, err := svc.CheckPermission(context.Background(), []string{"user"}, "user.delete")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if hasPerm {
			t.Error("Expected to NOT have permission, got true")
		}
	})

	t.Run("Cache Get Failed", func(t *testing.T) {
		mockCache := &mockRBACCacheRepository{
			getFunc: func(ctx context.Context, roleName []string) ([]string, error) {
				return nil, errors.New("cache lookup error")
			},
		}
		svc := services.NewRBACService(mockCache, mockRepo)

		hasPerm, err := svc.CheckPermission(context.Background(), []string{"user"}, "user.read")
		if err == nil || err.Error() != "cache lookup error" {
			t.Errorf("Expected 'cache lookup error', got %v", err)
		}
		if hasPerm {
			t.Error("Expected hasPerm to be false on error")
		}
	})
}
