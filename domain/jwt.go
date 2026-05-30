package domain

import "github.com/golang-jwt/jwt/v5"

type JwtClaims struct {
	Email  string   `json:"email"`
	UserID int      `json:"user_id"`
	Roles  []string `json:"roles"`
	jwt.RegisteredClaims
}

type JWTService interface {
	SignToken(user User) (string, error)
	ParseToken(tokenString string) (UserRoles, error)
}
