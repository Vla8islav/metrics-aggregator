package main

import (
	"context"
	"fmt"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"go.uber.org/zap"
)

func initDB(ctx context.Context, currentConfig *config.OptionsServer, logger *zap.Logger) (domain.MetricRepository, error) {
	var db domain.MetricRepository
	var err error

	// Case 1
	if currentConfig.DatabaseDSN.BeenSet {
		db, err = repository.NewPostgresStorage(currentConfig, currentConfig.MigrationsFolder.Value)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metrics repository: %w", err)
		}
		return db, nil
	}

	// Case 2
	if !currentConfig.Restore.BeenSet || !currentConfig.Restore.Value {
		logger.Info("making a fresh in-memory DB because the restore flag is false")
		db = repository.NewMemStorage(currentConfig)
		return db, nil
	}

	// Case 3
	if currentConfig.Restore.Value {
		logger.Info("trying to load data from file into the in-memory DB")
		dbInMemory := repository.NewMemStorage(currentConfig)
		go dbInMemory.RunSaver(ctx)
		err = dbInMemory.Restore(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to restore database")
		}
		db = dbInMemory
		return db, nil
	}

	return nil, fmt.Errorf("something strange happened: " +
		"restore and connection string parameters are incorrect")
}
