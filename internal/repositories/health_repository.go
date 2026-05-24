package repositories

import (
	"auth/domain"
	"context"
	"database/sql"
)

type healthRepository struct {
	db *sql.DB
}

func NewHealthRepository(db *sql.DB) domain.HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) PingContext(ctx context.Context) error {
	return r.db.PingContext(ctx)
}

func (r *healthRepository) GetStats(ctx context.Context) sql.DBStats {
	return r.db.Stats()
}
