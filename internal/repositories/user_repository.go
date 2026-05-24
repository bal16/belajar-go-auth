package repositories

import (
	"auth/domain"
	"context"

	"github.com/doug-martin/goqu"
)

type userRepository struct {
	db *goqu.Database
}

func (r *userRepository) FindByEmailWithLocalAuth(ctx context.Context, email string) (domain.UserEmailAuth, error) {
	var user domain.UserEmailAuth

	_, err := r.db.From(goqu.I("users").As("u")).
		Join(
			goqu.I("user_authentications").As("ua"),
			goqu.On(goqu.Ex{"u.email": goqu.I("ua.provider_key")}),
		).
		Where(goqu.Ex{
			"ua.provider":     "local",
			"ua.provider_key": email,
			"u.email":         email,
		}).
		Select(
			goqu.I("u.id").As("id"),
			goqu.I("u.email").As("email"),
			goqu.I("ua.password_hash").As("password_hash"),
		).
		ScanStructContext(ctx, &user)

	if err != nil {
		return domain.UserEmailAuth{}, err
	}

	return user, nil
}

func NewUser(db *goqu.Database) domain.UserRepository {
	return &userRepository{
		db,
	}
}
