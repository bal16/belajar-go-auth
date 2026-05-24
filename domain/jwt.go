package domain

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	Email string `json:"email"`
	UserID   int   `json:"user_id"`
	jwt.RegisteredClaims
}

type JWTService interface {
	SignToken(user User) (string, error)
	ParseToken(tokenString string) (int, error)
}