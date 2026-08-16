package repository

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

func getTestStorage(t *testing.T) *PostgresStorage {
	t.Helper()
	os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true")

	ctx := context.Background()

	pgContainer, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("metrics_db"),
		tcpostgres.WithUsername("default_user"),
		tcpostgres.WithPassword("default_password"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	require.NoError(t, err, "failed to start postgres container")
	t.Cleanup(func() {
		_ = pgContainer.Terminate(ctx)
	})

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	cfg, err := config.ReadFlagsServer([]string{}, zap.NewNop())
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	cfg.DatabaseDSN.Value = dsn
	cfg.DatabaseDSN.BeenSet = true

	testStorage, err := NewPostgresStorage(cfg, "../../migrations")
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	return testStorage
}
