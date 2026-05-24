package domain

import (
	"context"
	"time"

	"auth/dto"
)

type UserAuth struct {
	ID          int    `db:"id"`
	UserID      int    `db:"user_id"`
	Provider    string `db:"provider"`
	ProviderKey string `db:"provider_key"`
	Password    string `db:"password_hash" json:"-"`
}

type UserRefreshToken struct {
	ID        string    `db:"id"`
	UserID    int       `db:"user_id"`
	Token     string    `db:"token"`
	IsRevoked bool      `db:"is_revoked"`
	ExpiresAt time.Time `db:"expires_at"`
}

type UserEmailAuth struct {
	User
	Password string `db:"password_hash" json:"-"`
}

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (string, string, error)
	Register(ctx context.Context, req dto.RegisterRequest) error
}
