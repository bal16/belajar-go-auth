package services_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"auth/domain"
	"auth/dto"
	"auth/internal/services"

	"golang.org/x/crypto/bcrypt"
)

type mockedData struct {
	user             domain.User
	userWithPassword domain.UserEmailAuth
}

func getMockedData() mockedData {
	baseUser := domain.User{
		ID:    1,
		Email: "test@example.com",
		Name:  "Test User",
	}
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("secret123"), bcrypt.DefaultCost)

	return mockedData{
		user: baseUser,
		userWithPassword: domain.UserEmailAuth{
			User:     baseUser,
			Password: string(hashedPassword),
		},
	}
}

// --- Mock UserRepository ---

type mockUserRepository struct {
	findByEmailWithLocalAuthFunc func(ctx context.Context, email string) (domain.UserEmailAuth, error)
	createWithLocalAuthFunc      func(ctx context.Context, user domain.UserEmailAuth) error
	createRefreshTokenFunc       func(ctx context.Context, token domain.UserRefreshToken) error
	findByIDFunc                 func(ctx context.Context, id int) (domain.User, error)
}

func (m *mockUserRepository) FindByEmailWithLocalAuth(ctx context.Context, email string) (domain.UserEmailAuth, error) {
	return m.findByEmailWithLocalAuthFunc(ctx, email)
}

func (m *mockUserRepository) CreateWithLocalAuth(ctx context.Context, user domain.UserEmailAuth) error {
	return m.createWithLocalAuthFunc(ctx, user)
}

func (m *mockUserRepository) CreateRefreshToken(ctx context.Context, token domain.UserRefreshToken) error {
	if m.createRefreshTokenFunc != nil {
		return m.createRefreshTokenFunc(ctx, token)
	}
	return nil
}

func (m *mockUserRepository) FindByID(ctx context.Context, id int) (domain.User, error) {
	return m.findByIDFunc(ctx, id)
}

// --- Mock JWTService ---

type mockJWTService struct {
	signTokenFunc  func(user domain.User) (string, error)
	parseTokenFunc func(tokenString string) (int, error)
}

func (m *mockJWTService) SignToken(user domain.User) (string, error) {
	return m.signTokenFunc(user)
}

func (m *mockJWTService) ParseToken(tokenString string) (int, error) {
	return m.parseTokenFunc(tokenString)
}

// --- Tests ---

func TestAuthService_Login(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return getMockedData().userWithPassword, nil
			},
			createRefreshTokenFunc: func(ctx context.Context, token domain.UserRefreshToken) error {
				return nil
			},
		}
		mockJWT := &mockJWTService{
			signTokenFunc: func(user domain.User) (string, error) {
				return "mocked.jwt.token", nil
			},
		}
		service := services.NewAuthService(mockRepo, mockJWT)

		accessToken, refreshToken, err := service.Login(context.Background(), dto.LoginRequest{Email: "test@example.com", Password: "secret123"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if accessToken != "mocked.jwt.token" {
			t.Errorf("Expected token 'mocked.jwt.token', got %v", accessToken)
		}
		parts := strings.Split(refreshToken, ".")
		if len(parts) != 2 || len(parts[1]) != 64 {
			t.Errorf("Expected refresh token format 'id.token' with token length 64, got %s", refreshToken)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return domain.UserEmailAuth{}, errors.New("db error not found")
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		_, _, err := service.Login(context.Background(), dto.LoginRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "invalid username or password" {
			t.Errorf("Expected 'invalid username or password' error, got %v", err)
		}
	})

	t.Run("Invalid Password", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return getMockedData().userWithPassword, nil
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		_, _, err := service.Login(context.Background(), dto.LoginRequest{Email: "test@example.com", Password: "wrongpassword"})
		if err == nil || err.Error() != "invalid username or password" {
			t.Errorf("Expected 'invalid username or password' error, got %v", err)
		}
	})

	t.Run("Token Generation Failed", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return getMockedData().userWithPassword, nil
			},
		}
		mockJWT := &mockJWTService{
			signTokenFunc: func(user domain.User) (string, error) {
				return "", errors.New("jwt signing error")
			},
		}
		service := services.NewAuthService(mockRepo, mockJWT)

		_, _, err := service.Login(context.Background(), dto.LoginRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "failed to generate access token" {
			t.Errorf("Expected 'failed to generate access token' error, got %v", err)
		}
	})

	t.Run("Refresh Token Save Failed", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return getMockedData().userWithPassword, nil
			},
			createRefreshTokenFunc: func(ctx context.Context, token domain.UserRefreshToken) error {
				return errors.New("db insert error")
			},
		}
		mockJWT := &mockJWTService{
			signTokenFunc: func(user domain.User) (string, error) {
				return "mocked.jwt.token", nil
			},
		}
		service := services.NewAuthService(mockRepo, mockJWT)

		_, _, err := service.Login(context.Background(), dto.LoginRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "failed to save refresh token" {
			t.Errorf("Expected 'failed to save refresh token' error, got %v", err)
		}
	})
}

func TestAuthService_Register(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			createWithLocalAuthFunc: func(ctx context.Context, user domain.UserEmailAuth) error {
				return nil
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		err := service.Register(context.Background(), dto.RegisterRequest{Name: "Test", Email: "test@example.com", Password: "secret123"})
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
	})

	t.Run("Hash Password Failed", func(t *testing.T) {
		mockRepo := &mockUserRepository{}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		// bcrypt returns an error if the password length exceeds 72 bytes.
		longPassword := "01234567890123456789012345678901234567890123456789012345678901234567890123" // 74 chars
		err := service.Register(context.Background(), dto.RegisterRequest{Name: "Test", Email: "test@example.com", Password: longPassword})
		if err == nil || err.Error() != "failed to hash password" {
			t.Errorf("Expected 'failed to hash password' error, got %v", err)
		}
	})

	t.Run("Create Failed - User Already Exists", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			createWithLocalAuthFunc: func(ctx context.Context, user domain.UserEmailAuth) error {
				return errors.New("db insert error")
			},
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return domain.UserEmailAuth{
					User: domain.User{ID: 1, Email: "test@example.com"},
				}, nil
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		err := service.Register(context.Background(), dto.RegisterRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "user already exists" {
			t.Errorf("Expected 'user already exists' error, got %v", err)
		}
	})

	t.Run("Create Failed - DB Error on FindByEmail", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			createWithLocalAuthFunc: func(ctx context.Context, user domain.UserEmailAuth) error {
				return errors.New("db insert error")
			},
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return domain.UserEmailAuth{}, errors.New("db connection lost")
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		err := service.Register(context.Background(), dto.RegisterRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "Something Happen with database" {
			t.Errorf("Expected 'Something Happen with database' error, got %v", err)
		}
	})

	t.Run("Create Failed - Unknown Reason", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			createWithLocalAuthFunc: func(ctx context.Context, user domain.UserEmailAuth) error {
				return errors.New("db insert error")
			},
			findByEmailWithLocalAuthFunc: func(ctx context.Context, email string) (domain.UserEmailAuth, error) {
				return domain.UserEmailAuth{}, nil
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		err := service.Register(context.Background(), dto.RegisterRequest{Email: "test@example.com", Password: "secret123"})
		if err == nil || err.Error() != "Failed to create user. Unknown error" {
			t.Errorf("Expected 'Failed to create user. Unknown error' error, got %v", err)
		}
	})
}

func TestAuthService_GetMe(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByIDFunc: func(ctx context.Context, id int) (domain.User, error) {
				return getMockedData().user, nil
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		user, err := service.GetMe(context.Background(), 1)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if user.ID != 1 || user.Email != "test@example.com" || user.Name != "Test User" {
			t.Errorf("Expected user ID 1, email test@example.com, and name Test User, got %+v", user)
		}
	})

	t.Run("User Not Found", func(t *testing.T) {
		mockRepo := &mockUserRepository{
			findByIDFunc: func(ctx context.Context, id int) (domain.User, error) {
				return domain.User{}, errors.New("db error not found")
			},
		}
		service := services.NewAuthService(mockRepo, &mockJWTService{})

		_, err := service.GetMe(context.Background(), 99)
		if err == nil {
			t.Error("Expected error, got nil")
		}
		if err != nil && err.Error() != "user not found" {
			t.Errorf("Expected 'user not found' error, got %v", err.Error())
		}
	})
}
