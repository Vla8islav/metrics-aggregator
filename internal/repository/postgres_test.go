package repository

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
)

var testStorage *PostgresStorage

func TestMain(m *testing.M) {
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
	if err != nil {
		log.Fatalf("failed to start postgres container: %v", err)
	}

	defer func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}()

	dsn, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalf("failed to get connection string: %v", err)
	}

	cfg := config.ReadFlags([]string{})
	cfg.DatabaseDSN.Value = dsn
	cfg.DatabaseDSN.BeenSet = true

	testStorage, err = NewPostgresStorage(cfg, "../../migrations")
	if err != nil {
		log.Fatalf("failed to create storage: %v", err)
	}

	code := m.Run()
	os.Exit(code)
}
