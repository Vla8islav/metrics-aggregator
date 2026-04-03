package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	models "github.com/Vla8islav/metrics-aggregator/internal/model"
)

type PostgresStorage struct {
	config *config.Options
	db     *sql.DB
}

func NewPostgresStorage(config *config.Options) *PostgresStorage {
	if config == nil || !config.DatabaseDSN.BeenSet {
		log.Fatalf("config is nil")
	}
	if !config.DatabaseDSN.BeenSet {
		log.Fatalf("Database DSN wasn't set")
	}

	dsn := config.DatabaseDSN.Value
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	// verify conntection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		log.Fatal("couldn't ping postgres db " + err.Error())
	}

	return &PostgresStorage{config: config, db: db}
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

	rows, err := s.db.QueryContext(ctx, "SELECT * FROM metric_counters")

	return models.MetricsExport{
		Counters: counters,
		Gauges:   gauges,
	}, nil

}

func (s *PostgresStorage) IncrementCounter(ctx context.Context, name string, number int64) error {

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	if _, ok := s.namedCounter[name]; ok {
		s.namedCounter[name] = s.namedCounter[name] + number
	} else {
		s.namedCounter[name] = number
	}
	s.mu.Unlock()
	err := s.saveIfImmediateSaveIsSet(ctx)
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
	s.mu.Lock()
	s.namedGauge[name] = gauge
	s.mu.Unlock()
	err := s.saveIfImmediateSaveIsSet(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStorage) GetGauge(ctx context.Context, name string) (float64, error) {
	select {
	case <-ctx.Done():
		return 0.0, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedGauge[name]
	if !ok {
		return 0.0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

func (s *PostgresStorage) GetCounter(ctx context.Context, name string) (int64, error) {
	select {
	case <-ctx.Done():
		return 0.0, ctx.Err()
	default:
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.namedCounter[name]
	if !ok {
		return 0, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	return value, nil
}

func (s *PostgresStorage) SaveState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	export, err := s.GetAll(ctx)
	if err != nil {
		return err
	}

	filename := s.config.FileStoragePath
	file, err := os.OpenFile(filename.Value, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	marshalingResult, err := json.Marshal(export)
	if err != nil {
		return err
	}
	_, err = file.Write(marshalingResult)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresStorage) LoadState(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// check if file exists
	if _, err := os.Stat(s.config.FileStoragePath.Value); os.IsNotExist(err) {
		return nil
	}

	// now we know it does, so let's load it
	file, err := os.OpenFile(s.config.FileStoragePath.Value, os.O_RDONLY, 0644)
	if file != nil {
		defer file.Close()
	}
	if err != nil {
		return err
	}

	fileContent, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var metrics models.MetricsExport

	err = json.Unmarshal(fileContent, &metrics)
	if err != nil {
		return err
	}
	if metrics.Counters == nil {
		metrics.Counters = make(map[string]int64)
	}
	if metrics.Gauges == nil {
		metrics.Gauges = make(map[string]float64)
	}

	s.namedCounter = metrics.Counters
	s.namedGauge = metrics.Gauges

	return nil
}
