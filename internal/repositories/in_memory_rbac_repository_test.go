package repositories_test

import (
	"context"
	"reflect"
	"testing"

	"auth/domain"
	"auth/internal/repositories"
)

func TestInMemoryRBACRepository(t *testing.T) {
	repo := repositories.NewInMemoryRBACRepository()
	ctx := context.Background()

	t.Run("Get Empty Cache", func(t *testing.T) {
		perms, err := repo.Get(ctx, []string{"admin"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(perms) != 0 {
			t.Errorf("Expected 0 permissions, got %d", len(perms))
		}
	})

	// Setup mock data for Set
	mockRolePerms := []domain.RolePermissions{
		{
			Role: domain.Role{ID: 1, Name: "admin"},
			Permissions: []domain.Permission{
				{ID: 1, Name: "user.read"},
				{ID: 2, Name: "user.write"},
			},
		},
		{
			Role: domain.Role{ID: 2, Name: "user"},
			Permissions: []domain.Permission{
				{ID: 1, Name: "user.read"},
			},
		},
	}

	t.Run("Set Cache", func(t *testing.T) {
		err := repo.Set(ctx, mockRolePerms)
		if err != nil {
			t.Fatalf("Expected no error on Set, got %v", err)
		}
	})

	t.Run("Get Single Role", func(t *testing.T) {
		perms, err := repo.Get(ctx, []string{"user"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		expected := []string{"user.read"}
		if !reflect.DeepEqual(perms, expected) {
			t.Errorf("Expected %v, got %v", expected, perms)
		}
	})

	t.Run("Get Multiple Roles", func(t *testing.T) {
		perms, err := repo.Get(ctx, []string{"admin", "user"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Based on the current implementation, permissions will be appended sequentially.
		// admin permissions first, then user permissions.
		expected := []string{"user.read", "user.write", "user.read"}
		if !reflect.DeepEqual(perms, expected) {
			t.Errorf("Expected %v, got %v", expected, perms)
		}
	})

	t.Run("Get Non-Existent Role", func(t *testing.T) {
		perms, err := repo.Get(ctx, []string{"guest"})
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if len(perms) != 0 {
			t.Errorf("Expected 0 permissions, got %d", len(perms))
		}
	})
}
