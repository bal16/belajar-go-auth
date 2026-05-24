package repositories_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"auth/domain"
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

func TestUserRepository_CreateWithLocalAuth(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	newUser := domain.UserEmailAuth{
		User: domain.User{
			Email: "test@example.com",
			Name:  "Test User",
		},
		Password: "secretpassword",
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.CreateWithLocalAuth(context.Background(), newUser)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Failure - Duplicate Email", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnError(errors.New("unique constraint violation"))
		mock.ExpectRollback()

		err := repo.CreateWithLocalAuth(context.Background(), newUser)
		if err == nil || err.Error() != "email already exists" {
			t.Errorf("Expected 'email already exists', got %v", err)
		}
	})

	t.Run("Failure - Insert Auth Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnError(errors.New("insert auth error"))
		mock.ExpectRollback()

		err := repo.CreateWithLocalAuth(context.Background(), newUser)
		if err == nil || err.Error() != "insert auth error" {
			t.Errorf("Expected 'insert auth error', got %v", err)
		}
	})

	t.Run("Failure - Commit Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		err := repo.CreateWithLocalAuth(context.Background(), newUser)
		if err == nil || err.Error() != "commit error" {
			t.Errorf("Expected 'commit error', got %v", err)
		}
	})

	t.Run("Failure - Begin Transaction", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		err := repo.CreateWithLocalAuth(context.Background(), newUser)
		if err == nil || err.Error() != "begin error" {
			t.Errorf("Expected 'begin error', got %v", err)
		}
	})
}

func TestUserRepository_CreateRefreshToken(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	token := domain.UserRefreshToken{
		UserID:    1,
		Token:     "test-token",
		IsRevoked: false,
		ExpiresAt: time.Now(),
	}

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_refresh_tokens".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateRefreshToken(context.Background(), token)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("insert error")
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_refresh_tokens".*`).
			WillReturnError(expectedErr)

		err := repo.CreateRefreshToken(context.Background(), token)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "email", "name"}).
			AddRow(1, "test@example.com", "Test User")

		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*WHERE.*"id".*1.*`).
			WillReturnRows(rows)

		user, err := repo.FindByID(context.Background(), 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.ID != 1 {
			t.Errorf("Expected user ID 1, got %v", user.ID)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("db query error")
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*WHERE.*"id".*1.*`).
			WillReturnError(expectedErr)

		_, err := repo.FindByID(context.Background(), 1)
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("Unfulfilled expectations: %s", err)
		}
	})
}
