// Package repository package for working with data, mostly postgres
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

// PostgresStorage stores metrics in PostgreSQL and applies database migrations on startup
type PostgresStorage struct {
	config     *config.OptionsServer
	db         *sql.DB
	classifier *PostgresErrorClassifier
}

// isRetriablePostgresError reports whether err is a retryable PostgreSQL error
func (s *PostgresStorage) isRetriablePostgresError(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if s.classifier == nil {
		return false
	}

	switch s.classifier.Classify(pgErr) {
	case Retriable:
		return true
	case NonRetriable:
		return false
	default:
		return false
	}
}

// NewPostgresStorage opens a PostgreSQL connection, verifies it, and runs pending migrations
func NewPostgresStorage(config *config.OptionsServer, migrationsFolder string) (*PostgresStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("config is nil")
	}
	if !config.DatabaseDSN.BeenSet {
		return nil, fmt.Errorf("database DSN wasn't set")
	}

	dsn := config.DatabaseDSN.Value
	db, err := sql.Open("pgx", dsn)

	if err != nil {
		return nil, err
	}
	// Configure connection pool limits.
	db.SetConnMaxIdleTime(time.Minute * 2)
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)

	storage := PostgresStorage{config: config, db: db, classifier: NewPostgresErrorClassifier()}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = storage.Ping(ctx)
	if err != nil {
		if closeError := storage.db.Close(); closeError != nil {
			return nil, fmt.Errorf("failed to ping postgres %w also failed to close db %v", err, closeError)
		}
		return nil, fmt.Errorf("failed to ping postgres %w", err)
	}

	// Run all pending migrations from migrationsFolder.
	if err := goose.Up(db, migrationsFolder); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply goose migrations: %w", err)
	}

	return &storage, nil
}

// Ping checks whether the PostgreSQL connection is available
func (s *PostgresStorage) Ping(ctx context.Context) error {

	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("couldn't ping postgres db: %w", err)
	}

	return nil
}

// Restore verifies the persisted metric state when restore is enabled
func (s *PostgresStorage) Restore(ctx context.Context) error {
	if s.config.Restore.Value {
		err := s.LoadState(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAll returns all stored counters and gauges
func (s *PostgresStorage) GetAll(ctx context.Context) (models.MetricsExport, error) {
	select {
	case <-ctx.Done():
		return models.MetricsExport{}, ctx.Err()
	default:
	}

	counters, err := s.getCounters(ctx)
	if err != nil {
		return models.MetricsExport{}, err
	}

	gauges, err := s.getGauges(ctx)
	if err != nil {
		return models.MetricsExport{}, err
	}

	return models.MetricsExport{
		Counters: counters,
		Gauges:   gauges,
	}, nil

}

// getCounters returns all stored counters
func (s *PostgresStorage) getCounters(ctx context.Context) (map[string]int64, error) {
	countersFinal := make(map[string]int64)
	err := s.withRetry(ctx, func() error {
		counters := make(map[string]int64)
		rows, err := s.db.QueryContext(ctx, "SELECT name, value FROM metric_counters")
		if err != nil {
			return fmt.Errorf("querying counters: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			var value int64

			if err = rows.Scan(&name, &value); err != nil {
				return err
			}
			counters[name] = value
		}
		if err = rows.Err(); err != nil {
			return err
		}
		countersFinal = counters
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("querying counters: %w", err)
	}

	return countersFinal, nil
}

// getGauges returns all stored gauges
func (s *PostgresStorage) getGauges(ctx context.Context) (map[string]float64, error) {
	gaugesFinal := make(map[string]float64)
	err := s.withRetry(ctx, func() error {
		gauges := make(map[string]float64)
		rows, err := s.db.QueryContext(ctx, "SELECT name, value FROM metric_gauges")
		if err != nil {
			return fmt.Errorf("querying gauges: %w", err)
		}
		defer rows.Close()

		for rows.Next() {
			var name string
			var value float64

			if err = rows.Scan(&name, &value); err != nil {
				return err
			}
			gauges[name] = value
		}
		if err = rows.Err(); err != nil {
			return err
		}
		gaugesFinal = gauges
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("querying gauges: %w", err)
	}

	return gaugesFinal, nil
}

// IncrementCounter adds number to the named counter
func (s *PostgresStorage) IncrementCounter(ctx context.Context, name string, number int64) error {
	return s.withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx, `
	    INSERT INTO metric_counters (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = metric_counters.value + EXCLUDED.value
	`, name, number)
		if err != nil {
			return fmt.Errorf("increment counter failed %q: %w", name, err)
		}
		return nil
	})
}

// SetGauge updates the current value of the named gauge metric
func (s *PostgresStorage) SetGauge(ctx context.Context, name string, gauge float64) error {
	return s.withRetry(ctx, func() error {
		_, err := s.db.ExecContext(ctx, `
	    INSERT INTO metric_gauges (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = EXCLUDED.value
	`, name, gauge)
		if err != nil {
			return fmt.Errorf("set gauge failed %q: %w", name, err)
		}
		return nil
	})
}

// batchSetGauge stores multiple gauge values in one retryable transaction
func (s *PostgresStorage) batchSetGauge(ctx context.Context, names []string, gauges []float64) error {
	if len(names) == 0 {
		return nil
	}
	if len(gauges) != len(names) {
		return fmt.Errorf("batchSetGauge called with incorrect number of gauges")
	}
	positionalArguments := make([]string, len(names))
	var values []interface{}

	for i, name := range names {
		position1 := i*2 + 1
		position2 := i*2 + 2
		gaugeValue := gauges[i]

		positionalArguments[i] = fmt.Sprintf("($%d, $%d)", position1, position2)
		values = append(values, name, gaugeValue)
	}

	return s.withRetryTx(ctx,
		func(tx *sql.Tx) error { return s.batchSetGaugeTx(ctx, tx, positionalArguments, values) },
	)
}

// batchSetGaugeTx stores multiple gauge values within tx
func (s *PostgresStorage) batchSetGaugeTx(ctx context.Context,
	tx *sql.Tx,
	positionalArguments []string,
	values []interface{}) error {
	query := fmt.Sprintf(`INSERT INTO metric_gauges (name, value) 
	    VALUES %s
	    ON CONFLICT (name)
	    DO UPDATE SET value = EXCLUDED.value
	`, strings.Join(positionalArguments, ","))
	_, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("set gauge batch failed %v: %w", positionalArguments, err)
	}
	return nil
}

// batchIncrementCounters increments multiple counter values in one retryable tx
func (s *PostgresStorage) batchIncrementCounters(ctx context.Context, names []string, counters []int64) error {
	if len(names) == 0 {
		return nil
	}
	if len(counters) != len(names) {
		return fmt.Errorf("batchIncrementCounters called with incorrect number of counters")
	}
	positionalArguments := make([]string, len(names))
	var values []interface{}

	for i, name := range names {
		position1 := i*2 + 1
		position2 := i*2 + 2
		counterIncrement := counters[i]

		positionalArguments[i] = fmt.Sprintf("($%d, $%d)", position1, position2)
		values = append(values, name, counterIncrement)
	}

	return s.withRetryTx(ctx,
		func(tx *sql.Tx) error { return s.batchIncrementCountersTx(ctx, tx, positionalArguments, values) },
	)
}

// batchIncrementCountersTx increments multiple counter values within tx
func (s *PostgresStorage) batchIncrementCountersTx(ctx context.Context,
	tx *sql.Tx,
	positionalArguments []string,
	values []interface{}) error {
	query := fmt.Sprintf(`INSERT INTO metric_counters (name, value) 
	    VALUES %s
	    ON CONFLICT (name)
	    DO UPDATE SET value = metric_counters.value + EXCLUDED.value
	`, strings.Join(positionalArguments, ","))
	_, err := tx.ExecContext(ctx, query, values...)
	if err != nil {
		return fmt.Errorf("increment counter batch failed %v: %w", positionalArguments, err)
	}
	return nil
}

// GetGauge returns the current value of the named gauge metric
func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := s.withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, `SELECT value FROM metric_gauges WHERE name = $1`, name).Scan(&value)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}

	if err != nil {
		return 0, fmt.Errorf("getting gauge failed %s: %w", name, err)
	}

	return value, nil
}

// GetCounter returns the current value of the named counter
func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	err := s.withRetry(ctx, func() error {
		return s.db.QueryRowContext(ctx, `SELECT value FROM metric_counters WHERE name = $1`, name).Scan(&value)
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	} else if err != nil {
		return 0, fmt.Errorf("getting counter failed %s: %w", name, err)
	}

	return value, nil

}

// LoadState verifies that persisted metric state is available from PostgreSQL
func (s *PostgresStorage) LoadState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// PostgreSQL already owns the persisted state, so no explicit load is needed

	return nil
}

// UpdateMetrics applies a batch of metric updates
func (s *PostgresStorage) UpdateMetrics(ctx context.Context, input []models.Metrics) error {
	if len(input) == 0 {
		return nil
	}

	for _, m := range input {
		if m.MType == models.Gauge && m.Value != nil {
			err := s.SetGauge(ctx, m.ID, *m.Value)
			if err != nil {
				return err
			}
		}
		if m.MType == models.Counter && m.Delta != nil {
			err := s.IncrementCounter(ctx, m.ID, *m.Delta)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
