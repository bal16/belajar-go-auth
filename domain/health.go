package domain

import (
	"context"
	"database/sql"
)

type HealthService interface {
	Health(ctx context.Context) map[string]string
}

type HealthRepository interface {
	PingContext(ctx context.Context) error
	GetStats(ctx context.Context) sql.DBStats
}