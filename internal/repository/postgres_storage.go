package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	goose "github.com/pressly/goose/v3"
)

type PostgresStorage struct {
	config     *config.Options
	db         *sql.DB
	classifier *PostgresErrorClassifier
}

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

func NewPostgresStorage(config *config.Options, migrationsFolder string) (*PostgresStorage, error) {
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

	// Run all pending migrations from migrations/
	if err := goose.Up(db, migrationsFolder); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("apply goose migrations: %w", err)
	}

	return &storage, nil
}

func (s *PostgresStorage) Ping(ctx context.Context) error {

	if s.db == nil {
		return errors.New("database pointer was nil")
	}

	// verify connection
	if err := s.db.PingContext(ctx); err != nil {
		return fmt.Errorf("couldn't ping postgres db: %w", err)
	}

	return nil
}

func (s *PostgresStorage) TruncateEverything() error {
	if s.db == nil {
		return errors.New("database pointer is nil")
	}
	_, err := s.db.Exec(`
		TRUNCATE TABLE metric_counters, metric_gauges RESTART IDENTITY CASCADE
	`)
	if err != nil {
		return fmt.Errorf("truncate metric tables: %w", err)
	}
	return nil
}

func (s *PostgresStorage) Restore(ctx context.Context) error {
	if s.config.Restore.Value {
		err := s.LoadState(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) RunSaver(_ context.Context) error {
	return nil
}

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

func (s *PostgresStorage) getCounters(ctx context.Context) (map[string]int64, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT name, value FROM metric_counters")
	if err != nil {
		if s.isRetriablePostgresError(err) {
			return nil, fmt.Errorf("getting all counters failed retryable: %w", err)
		}

		return nil, fmt.Errorf("getting all counters failed: %w", err)
	}
	defer rows.Close()

	counters := make(map[string]int64)
	for rows.Next() {
		var name string
		var value int64

		if err = rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		counters[name] = value
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return counters, nil
}

func (s *PostgresStorage) getGauges(ctx context.Context) (map[string]float64, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT name, value FROM metric_gauges")
	if err != nil {
		if s.isRetriablePostgresError(err) {
			return nil, fmt.Errorf("getting all gauges failed retryable: %w", err)
		}

		return nil, fmt.Errorf("getting all gauges failed: %w", err)
	}
	defer rows.Close()

	gauges := make(map[string]float64)
	for rows.Next() {
		var name string
		var value float64

		if err = rows.Scan(&name, &value); err != nil {
			return nil, err
		}
		gauges[name] = value
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return gauges, nil
}

func (s *PostgresStorage) IncrementCounter(ctx context.Context, name string, number int64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (sql.Result, error) {
		return s.db.ExecContext(ctx, `
	    INSERT INTO metric_counters (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = metric_counters.value + EXCLUDED.value
	`, name, number)
	})
	if err != nil {
		return fmt.Errorf("increment counter failed %s: %w", name, err)
	}
	return nil
}

func (s *PostgresStorage) SetGauge(ctx context.Context, name string, gauge float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (sql.Result, error) {
		return s.db.ExecContext(ctx, `
	    INSERT INTO metric_gauges (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = EXCLUDED.value
	`, name, gauge)
	})
	if err != nil {
		return fmt.Errorf("set gauge failed %s: %w", name, err)
	}
	return nil
}

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

	query := fmt.Sprintf(`INSERT INTO metric_gauges (name, value) 
	    VALUES %s
	    ON CONFLICT (name)
	    DO UPDATE SET value = EXCLUDED.value
	`, strings.Join(positionalArguments, ","))

	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (sql.Result, error) {
		return s.db.ExecContext(ctx, query, values...)
	})
	if err != nil {
		return fmt.Errorf("set gauge batch failed %v: %w", names, err)
	}
	return nil
}

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

	query := fmt.Sprintf(`INSERT INTO metric_counters (name, value)
	    VALUES %s
	    ON CONFLICT (name)
	    DO UPDATE SET value = metric_counters.value + EXCLUDED.value
	`, strings.Join(positionalArguments, ","))

	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (sql.Result, error) {
		return s.db.ExecContext(ctx, query, values...)
	})
	if err != nil {
		return fmt.Errorf("increment counter failed %v: %w", names, err)
	}
	return nil
}
func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (struct{}, error) {
		err := s.db.QueryRowContext(ctx, "SELECT value FROM metric_gauges WHERE name = $1", name).Scan(&value)
		return struct{}{}, err
	})

	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	} else if err != nil {
		return 0, fmt.Errorf("getting gauge failed %s: %w", name, err)
	}

	return value, nil
}

func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	_, err := helpers.WithRetry(ctx, 3, s.isRetriablePostgresError, func() (struct{}, error) {
		err := s.db.QueryRowContext(ctx, "SELECT value FROM metric_counters WHERE name = $1", name).Scan(&value)
		return struct{}{}, err
	})
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	} else if err != nil {
		return 0, fmt.Errorf("getting counter failed %s: %w", name, err)
	}

	return value, nil

}

func (s *PostgresStorage) SaveState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// here save is handled by the DB

	return nil
}

func (s *PostgresStorage) LoadState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	// here load is handled by the DB

	return nil
}

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
