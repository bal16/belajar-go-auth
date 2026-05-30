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
		// Create a custom token to test both User and Roles parsing
		claims := &domain.JwtClaims{
			Email:  "test@example.com",
			UserID: 1,
			Roles:  []string{"admin", "user"},
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(conf.JWT.SECRET))

		userRoles, err := jwtSvc.ParseToken(tokenString)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if userRoles.User.ID != 1 {
			t.Errorf("Expected user ID 1, got %d", userRoles.User.ID)
		}
		if userRoles.User.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got %s", userRoles.User.Email)
		}
		if len(userRoles.Roles) != 2 || userRoles.Roles[0] != "admin" {
			t.Errorf("Expected roles [admin, user], got %v", userRoles.Roles)
		}
	})

	t.Run("Missing User Payload", func(t *testing.T) {
		// Create a token where UserID is 0
		claims := &domain.JwtClaims{
			Email:  "test@example.com",
			UserID: 0,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 15)),
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(conf.JWT.SECRET))

		_, err := jwtSvc.ParseToken(tokenString)
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

	t.Run("Expired Token", func(t *testing.T) {
		claims := &domain.JwtClaims{
			Email:  "test@example.com",
			UserID: 1,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-15 * time.Minute)), // Passed expiration
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString([]byte(conf.JWT.SECRET))

		_, err := jwtSvc.ParseToken(tokenString)
		if err == nil {
			t.Error("Expected error for expired token, got nil")
		}
	})

	t.Run("Malformed Token String", func(t *testing.T) {
		_, err := jwtSvc.ParseToken("not.a.valid.jwt.token")
		if err == nil {
			t.Error("Expected error for malformed token string, got nil")
		}
	})
}
