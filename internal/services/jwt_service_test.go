package services_test

import (
	"strings"
	"testing"
	"time"

	"auth/domain"
	"auth/internal/config"
	"auth/internal/services"

	"github.com/golang-jwt/jwt/v5"
)

// setupTestConfig returns a mock configuration for testing purposes.
func setupTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWT{
			SECRET: "my_super_secret_test_key",
			EXP:    15, // 15 minutes
		},
	}
}

func TestJWTService_SignToken(t *testing.T) {
	conf := setupTestConfig()
	jwtSvc := services.NewJWTService(conf)

	user := domain.User{
		ID:    1,
		Email: "test@example.com",
	}

	token, err := jwtSvc.SignToken(user)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected token to not be empty")
	}
}

func TestJWTService_ParseToken(t *testing.T) {
	conf := setupTestConfig()
	jwtSvc := services.NewJWTService(conf)

	t.Run("Valid Token", func(t *testing.T) {
		user := domain.User{
			ID:    1,
			Email: "test@example.com",
		}
		token, err := jwtSvc.SignToken(user)
		if err != nil {
			t.Fatalf("Failed to sign token: %v", err)
		}

		userID, err := jwtSvc.ParseToken(token)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if userID != user.ID {
			t.Errorf("Expected user ID %d, got %d", user.ID, userID)
		}
	})

	t.Run("Missing User Payload", func(t *testing.T) {
		// Create a token where UserID is 0
		user := domain.User{
			ID:    0,
			Email: "test@example.com",
		}
		token, _ := jwtSvc.SignToken(user)

		_, err := jwtSvc.ParseToken(token)
		if err == nil {
			t.Error("Expected error, got nil")
		} else if err.Error() != "invalid token: missing user payload" {
			t.Errorf("Expected 'invalid token: missing user payload', got %v", err)
		}
	})

	t.Run("Invalid Token Signature", func(t *testing.T) {
		user := domain.User{
			ID:    1,
			Email: "test@example.com",
		}
		token, _ := jwtSvc.SignToken(user)

		tamperedToken := token + "invalid_suffix"

		_, err := jwtSvc.ParseToken(tamperedToken)
		if err == nil {
			t.Error("Expected error for tampered token, got nil")
		}
	})

	t.Run("Unexpected Signing Method", func(t *testing.T) {
		// Create a token with 'none' signing method instead of HMAC
		claims := &domain.JwtClaims{
			Email:  "test@example.com",
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

		// We use jwt.UnsafeAllowNoneSignatureType purely for forcing a bad method in tests
		tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		_, err := jwtSvc.ParseToken(tokenString)
		if err == nil {
			t.Error("Expected error, got nil")
		} else if !strings.Contains(err.Error(), "unexpected signing method") {
			t.Errorf("Expected 'unexpected signing method' error, got %v", err)
		}
	})
}
