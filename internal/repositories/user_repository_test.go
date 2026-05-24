package repositories_test

import (
	"context"
	"errors"
	"testing"

	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/doug-martin/goqu"
)

// setupUserMockDB initializes a go-sqlmock instance and wraps it in a goqu Database
func setupUserMockDB(t *testing.T) (*goqu.Database, sqlmock.Sqlmock, func()) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	// Initialize goqu with postgres dialect to match your production database configuration
	goquDB := goqu.New("postgres", mockDB)

	cleanup := func() { mockDB.Close() }
	return goquDB, mock, cleanup
}

func TestUserRepository_FindByEmailWithLocalAuth(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "password_hash"}).
			AddRow(1, "test@example.com", "secret")

		// Use a case-insensitive regex to handle dynamically generated SQL by goqu
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnRows(rows)

		user, err := repo.FindByEmailWithLocalAuth(context.Background(), "test@example.com")

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("Expected user email 'test@example.com', got %v", user.Email)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("db query error")
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnError(expectedErr)

		_, err := repo.FindByEmailWithLocalAuth(context.Background(), "test@example.com")

		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
