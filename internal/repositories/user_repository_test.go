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

func TestUserRepository_RevokeRefreshToken(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	t.Run("Success", func(t *testing.T) {
		mock.ExpectExec(`(?i)UPDATE.*"user_refresh_tokens".*SET.*"is_revoked".*WHERE.*"id".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.RevokeRefreshToken(context.Background(), "test-token")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("update error")
		mock.ExpectExec(`(?i)UPDATE.*"user_refresh_tokens".*SET.*"is_revoked".*WHERE.*"id".*`).
			WillReturnError(expectedErr)

		err := repo.RevokeRefreshToken(context.Background(), "test-token")
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})
}

func TestUserRepository_FindRefreshToken(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	t.Run("Success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "token", "is_revoked", "expires_at"}).
			AddRow(1, 10, "test-token", false, time.Now())

		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_refresh_tokens".*WHERE.*"id".*`).
			WillReturnRows(rows)

		token, err := repo.FindRefreshToken(context.Background(), "test-token")
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if token.Token != "test-token" {
			t.Errorf("Expected token 'test-token', got %v", token.Token)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("query error")
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_refresh_tokens".*WHERE.*"id".*`).
			WillReturnError(expectedErr)

		_, err := repo.FindRefreshToken(context.Background(), "test-token")
		if err == nil || err.Error() != expectedErr.Error() {
			t.Errorf("Expected error %v, got %v", expectedErr, err)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_refresh_tokens".*WHERE.*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "token", "is_revoked", "expires_at"})) // empty rows

		_, err := repo.FindRefreshToken(context.Background(), "test-token")
		if err == nil || err.Error() != "refresh token not found or revoked" {
			t.Errorf("Expected error 'refresh token not found or revoked', got %v", err)
		}
	})
}

func TestUserRepository_FindOrCreateWithOAuth(t *testing.T) {
	db, mock, cleanup := setupUserMockDB(t)
	defer cleanup()
	repo := repositories.NewUser(db)

	oauthUser := domain.UserOauth{
		User: domain.User{
			Email: "oauth@example.com",
			Name:  "OAuth User",
		},
		Provider:    "google",
		ProviderKey: "google-123",
	}

	t.Run("Success - Auth Exists", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(10))
		mock.ExpectCommit()

		user, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.ID != 10 {
			t.Errorf("Expected user ID 10, got %v", user.ID)
		}
	})

	t.Run("Success - User Exists No Auth", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"}).AddRow(20, "oauth@example.com", "OAuth User"))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		user, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.ID != 20 {
			t.Errorf("Expected user ID 20, got %v", user.ID)
		}
	})

	t.Run("Success - User Not Exists", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"})) // empty
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(30))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		user, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.ID != 30 {
			t.Errorf("Expected user ID 30, got %v", user.ID)
		}
	})

	t.Run("Failure - Auth Query Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnError(errors.New("auth query error"))
		mock.ExpectRollback()

		_, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err == nil || err.Error() != "auth query error" {
			t.Errorf("Expected 'auth query error', got %v", err)
		}
	})

	t.Run("Failure - User Query Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnError(errors.New("user query error"))
		mock.ExpectRollback()

		_, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err == nil || err.Error() != "user query error" {
			t.Errorf("Expected 'user query error', got %v", err)
		}
	})

	t.Run("Failure - Insert User Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"})) // empty
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnError(errors.New("insert user error"))
		mock.ExpectRollback()

		_, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err == nil || err.Error() != "insert user error" {
			t.Errorf("Expected 'insert user error', got %v", err)
		}
	})

	t.Run("Failure - Insert Auth Error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"user_authentications".*`).
			WillReturnRows(sqlmock.NewRows([]string{"user_id"})) // empty
		mock.ExpectQuery(`(?i)SELECT.*FROM.*"users".*`).
			WillReturnRows(sqlmock.NewRows([]string{"id", "email", "name"})) // empty
		mock.ExpectQuery(`(?i)INSERT.*INTO.*"users".*RETURNING "id"`).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(30))
		mock.ExpectExec(`(?i)INSERT.*INTO.*"user_authentications".*`).
			WillReturnError(errors.New("insert auth error"))
		mock.ExpectRollback()

		_, err := repo.FindOrCreateWithOAuth(context.Background(), oauthUser)
		if err == nil || err.Error() != "insert auth error" {
			t.Errorf("Expected 'insert auth error', got %v", err)
		}
	})
}
