package services

import (
	"auth/domain"
	"auth/dto"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/nrednav/cuid2"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

// ValidateGoogleIDToken is extracted to a variable to allow mocking in tests.
var ValidateGoogleIDToken = idtoken.Validate

type authService struct {
	userRepo domain.UserRepository
	jwtSer   domain.JWTService
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (string, string, error) {
	user, err := s.userRepo.FindByEmailWithLocalAuth(ctx, req.Email)
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password))
	if err != nil {
		return "", "", errors.New("invalid username or password")
	}

	accessToken, err := s.jwtSer.SignToken(domain.User{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	})
	if err != nil {
		return "", "", errors.New("failed to generate access token")
	}

	refreshToken, err := makeRefreshToken()
	if err != nil {
		return "", "", err
	}

	refreshTokenData := domain.UserRefreshToken{
		ID:        cuid2.Generate(),
		UserID:    user.ID,
		Token:     refreshToken,
		IsRevoked: false,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}

	err = s.userRepo.CreateRefreshToken(ctx, refreshTokenData)
	if err != nil {
		return "", "", errors.New("failed to save refresh token")
	}

	validRefreshToken := refreshTokenData.ID + "." + refreshToken

	return accessToken, validRefreshToken, nil
}
func (s *authService) Register(ctx context.Context, req dto.RegisterRequest) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	user := domain.UserEmailAuth{
		User: domain.User{
			Email: req.Email,
			Name:  req.Name,
		},
		Password: string(hashedPassword),
	}

	err = s.userRepo.CreateWithLocalAuth(ctx, user)
	if err != nil {
		existingUser, err := s.userRepo.FindByEmailWithLocalAuth(ctx, user.Email)

		if err != nil {

			return errors.New("Something Happen with database")
		}

		if existingUser != (domain.UserEmailAuth{}) {
			return errors.New("user already exists")
		}

		return errors.New("Failed to create user. Unknown error")
	}

	return nil
}

func makeRefreshToken() (string, error) {
	refreshToken := make([]byte, 32)
	if _, err := rand.Read(refreshToken); err != nil {
		return "", errors.New("failed to generate refresh token")
	}

	return hex.EncodeToString(refreshToken), nil
}

func NewAuthService(userRepo domain.UserRepository, jwtSer domain.JWTService) domain.AuthService {
	return &authService{
		userRepo: userRepo,
		jwtSer:   jwtSer,
	}
}
