package services

import (
	"errors"
	"time"

	"auth/domain"
	"auth/internal/config"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
)

type jwtService struct {
	conf *config.Config
}

func (s *jwtService) SignToken(user domain.User) (string, error) {
	claims := &domain.JwtClaims{
		Email:  user.Email,
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * time.Duration(s.conf.JWT.EXP))),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.conf.JWT.SECRET))
}

func (s *jwtService) ParseToken(tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.JwtClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.conf.JWT.SECRET), nil
	})

	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*domain.JwtClaims); ok && token.Valid {
		if claims.UserID == 0 {
			return 0, errors.New("invalid token: missing user payload")
		}
		return claims.UserID, nil
	}

	return 0, errors.New("invalid token")
}

func NewJWTService(conf *config.Config) domain.JWTService {
	return &jwtService{
		conf: conf,
	}
}
