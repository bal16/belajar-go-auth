package repositories

import (
	"auth/domain"
	"context"
	"errors"
	"strings"

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

func (r *userRepository) CreateWithLocalAuth(ctx context.Context, user domain.UserEmailAuth) error {
	tx, err := r.db.Begin()

	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.From(goqu.I("users")).
		Returning("id").
		Insert(goqu.Record{
			"email": user.Email,
			"name":  user.Name,
		}).
		ScanValContext(ctx, &user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "unique constraint") {
			return errors.New("email already exists")
		}
		return err
	}

	_, err = tx.From("user_authentications").
		Insert(goqu.Record{
			"user_id":       user.ID,
			"provider":      "local",
			"provider_key":  user.Email,
			"password_hash": user.Password,
		}).
		ExecContext(ctx)

	if err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}
	return nil
}

func (r *userRepository) CreateRefreshToken(ctx context.Context, data domain.UserRefreshToken) error {
	_, err := r.db.From(goqu.I("user_refresh_tokens")).
		Insert(goqu.Record{
			"id":         data.ID,
			"user_id":    data.UserID,
			"token":      data.Token,
			"is_revoked": data.IsRevoked,
			"expires_at": data.ExpiresAt,
		}).
		ExecContext(ctx)

	return err
}

func (r *userRepository) FindByID(ctx context.Context, id int) (domain.User, error) {
	var user domain.User

	_, err := r.db.From("users").
		Where(goqu.Ex{
			"id": id,
		}).
		ScanStructContext(ctx, &user)

	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (r *userRepository) FindOrCreateWithOAuth(ctx context.Context, user domain.UserOauth) (domain.UserOauth, error) {
	tx, err := r.db.Begin()

	if err != nil {
		return domain.UserOauth{}, err
	}
	defer tx.Rollback()

	var existingAuth domain.UserAuth
	foundAuth, err := tx.From(goqu.I("user_authentications")).
		Where(goqu.Ex{
			"provider":     user.Provider,
			"provider_key": user.ProviderKey,
		}).
		ScanStructContext(ctx, &existingAuth)

	if err != nil {
		return domain.UserOauth{}, err
	}

	if foundAuth {
		user.ID = existingAuth.UserID

	} else {
		var existingUser domain.User
		foundUser, err := tx.From("users").
			Where(goqu.Ex{
				"email": user.Email,
			}).
			Select(
				goqu.I("id"),
				goqu.I("email"),
				goqu.I("name"),
			).
			ScanStructContext(ctx, &existingUser)

		if err != nil {
			return domain.UserOauth{}, err
		}

		if !foundUser {
			_, err = tx.From(goqu.I("users")).
				Returning("id").
				Insert(goqu.Record{
					"email": user.Email,
					"name":  user.Name,
				}).
				ScanValContext(ctx, &user.ID)

			if err != nil {
				return domain.UserOauth{}, err
			}
		} else {
			user.ID = existingUser.ID
		}

		_, err = tx.From(goqu.I("user_authentications")).
			Insert(goqu.Record{
				"user_id":      user.ID,
				"provider":     user.Provider,
				"provider_key": user.ProviderKey,
			}).
			ExecContext(ctx)

		if err != nil {
			return domain.UserOauth{}, err
		}
	}

	if err = tx.Commit(); err != nil {
		return domain.UserOauth{}, err
	}
	return user, nil
}

func NewUser(db *goqu.Database) domain.UserRepository {
	return &userRepository{
		db,
	}
}
