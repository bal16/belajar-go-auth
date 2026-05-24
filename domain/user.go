package domain

import (
	"context"
)

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

type UserRepository interface {
	FindByEmailWithLocalAuth(ctx context.Context, email string) (UserEmailAuth, error)
	CreateWithLocalAuth(ctx context.Context, user UserEmailAuth) error
	CreateRefreshToken(ctx context.Context, data UserRefreshToken) error
	FindByID(ctx context.Context, id int) (User, error)
	FindOrCreateWithOAuth(ctx context.Context, user UserOauth) (UserOauth, error)
}
