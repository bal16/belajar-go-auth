package repositories_test

import (
	"context"
	"errors"
	"testing"

	"auth/internal/repositories"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestHealthRepository_PingContext(t *testing.T) {
	// We must enable MonitorPingsOption to intercept PingContext calls
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repositories.NewHealthRepository(db)

	t.Run("Success", func(t *testing.T) {
		// Expect a ping and simulate a successful response (no error)
		mock.ExpectPing().WillReturnError(nil)

		err := repo.PingContext(context.Background())
		if err != nil {
			t.Errorf("expected no error, but got %v", err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		expectedErr := errors.New("db connection failed")
		// Expect a ping and simulate a failure
		mock.ExpectPing().WillReturnError(expectedErr)

		err := repo.PingContext(context.Background())
		if err == nil {
			t.Error("expected an error, but got nil")
		} else if err.Error() != expectedErr.Error() {
			t.Errorf("expected error %v, but got %v", expectedErr, err)
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("there were unfulfilled expectations: %s", err)
		}
	})
}

func TestHealthRepository_GetStats(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	repo := repositories.NewHealthRepository(db)

	// db.Stats() reads internal connection pool metrics. We can't mock the internal
	// stats with sqlmock, but we can verify the function executes and returns a valid struct.
	stats := repo.GetStats(context.Background())
	_ = stats // No specific assertions possible, just ensuring it doesn't panic
}