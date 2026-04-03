package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStorage struct {
	config *config.Options
	db     *sql.DB
}

func NewPostgresStorage(config *config.Options) (*PostgresStorage, error) {
	if config == nil || !config.DatabaseDSN.BeenSet {
		log.Fatalf("config is nil")
	}
	if !config.DatabaseDSN.BeenSet {
		log.Fatalf("Database DSN wasn't set")
	}

	dsn := config.DatabaseDSN.Value
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	storage := PostgresStorage{config: config, db: db}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = storage.Ping(ctx)
	if err != nil {
		return nil, err
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

func (s *PostgresStorage) Restore(ctx context.Context) error {
	if s.config.Restore.Value {
		err := s.LoadState(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) RunSaver(ctx context.Context) error {
	if s.config.Restore.Value && s.config.StoreInterval.Duration > 0 {
		fileSaveTicker := time.NewTicker(s.config.StoreInterval.Duration)
		defer fileSaveTicker.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case <-fileSaveTicker.C:
				err := s.SaveState(ctx)
				if err != nil {
					log.Printf("failed to save state: %v", err)
				}
			}
		}

	}
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
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM metric_counters")
	if err != nil {
		return nil, err
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
	rows, err := s.db.QueryContext(ctx, "SELECT * FROM metric_gauges")
	if err != nil {
		return nil, err
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
	_, err := s.db.ExecContext(ctx, `
	    INSERT INTO metric_counters (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = metric_counters.value + EXCLUDED.value
	`, name, number)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) saveIfImmediateSaveIsSet(ctx context.Context) error {
	if s.config.Restore.Value && s.config.StoreInterval.Duration == 0 {
		err := s.SaveState(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStorage) SetGauge(ctx context.Context, name string, gauge float64) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	_, err := s.db.ExecContext(ctx, `
	    INSERT INTO metric_gauges (name, value)
	    VALUES ($1, $2)
	    ON CONFLICT (name)
	    DO UPDATE SET value = EXCLUDED.value
	`, name, gauge)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	var value float64
	err := s.db.QueryRowContext(ctx, "SELECT * FROM metric_gauges WHERE name = $1", name).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return 0, err
	}

	return value, nil
}

func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	var value int64
	err := s.db.QueryRowContext(ctx, "SELECT * FROM metric_counters WHERE name = $1", name).Scan(&value)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
		}
		return 0, err
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
