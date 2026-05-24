package domain

import (
	"time"
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

type UserOauth struct {
	User
	Provider    string
	ProviderKey string
}

type AuthService interface {
}
