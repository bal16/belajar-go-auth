package services_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"auth/internal/services"
)

// --- Mock HealthRepository ---

type mockHealthRepository struct {
	pingContextFunc func(ctx context.Context) error
	getStatsFunc    func(ctx context.Context) sql.DBStats
}

func (m *mockHealthRepository) PingContext(ctx context.Context) error {
	if m.pingContextFunc != nil {
		return m.pingContextFunc(ctx)
	}
	return nil
}

func (m *mockHealthRepository) GetStats(ctx context.Context) sql.DBStats {
	if m.getStatsFunc != nil {
		return m.getStatsFunc(ctx)
	}
	return sql.DBStats{}
}

// --- Tests ---

func TestHealthService_Health(t *testing.T) {
	tests := []struct {
		name           string
		mockPing       func(ctx context.Context) error
		mockStats      func(ctx context.Context) sql.DBStats
		expectedStatus string
		expectedMsg    string
		expectedError  string
	}{
		{
			name: "Database Down",
			mockPing: func(ctx context.Context) error {
				return errors.New("connection refused")
			},
			expectedStatus: "down",
			expectedError:  "db down: connection refused",
		},
		{
			name: "Healthy Database",
			mockPing: func(ctx context.Context) error { return nil },
			mockStats: func(ctx context.Context) sql.DBStats {
				return sql.DBStats{OpenConnections: 10}
			},
			expectedStatus: "up",
			expectedMsg:    "It's healthy",
		},
		{
			name: "Heavy Load",
			mockPing: func(ctx context.Context) error { return nil },
			mockStats: func(ctx context.Context) sql.DBStats {
				return sql.DBStats{OpenConnections: 50}
			},
			expectedStatus: "up",
			expectedMsg:    "The database is experiencing heavy load.",
		},
		{
			name: "High Wait Events",
			mockPing: func(ctx context.Context) error { return nil },
			mockStats: func(ctx context.Context) sql.DBStats {
				return sql.DBStats{WaitCount: 1500}
			},
			expectedStatus: "up",
			expectedMsg:    "The database has a high number of wait events, indicating potential bottlenecks.",
		},
		{
			name: "Many Idle Connections Closed",
			mockPing: func(ctx context.Context) error { return nil },
			mockStats: func(ctx context.Context) sql.DBStats {
				return sql.DBStats{OpenConnections: 10, MaxIdleClosed: 6}
			},
			expectedStatus: "up",
			expectedMsg:    "Many idle connections are being closed, consider revising the connection pool settings.",
		},
		{
			name: "Many Connections Closed Due To Max Lifetime",
			mockPing: func(ctx context.Context) error { return nil },
			mockStats: func(ctx context.Context) sql.DBStats {
				return sql.DBStats{OpenConnections: 10, MaxLifetimeClosed: 6}
			},
			expectedStatus: "up",
			expectedMsg:    "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern.",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := &mockHealthRepository{
				pingContextFunc: tc.mockPing,
				getStatsFunc:    tc.mockStats,
			}
			service := services.NewHealthService(mockRepo)

			stats := service.Health(context.Background())

			if stats["status"] != tc.expectedStatus {
				t.Errorf("expected status %q, got %q", tc.expectedStatus, stats["status"])
			}

			if tc.expectedStatus == "down" {
				if stats["error"] != tc.expectedError {
					t.Errorf("expected error %q, got %q", tc.expectedError, stats["error"])
				}
			} else {
				if stats["message"] != tc.expectedMsg {
					t.Errorf("expected message %q, got %q", tc.expectedMsg, stats["message"])
				}
			}
		})
	}
}