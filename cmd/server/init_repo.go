package main

import (
	"context"

	"github.com/Vla8islav/metrics-aggregator/internal/config"
	"github.com/Vla8islav/metrics-aggregator/internal/domain"
	"github.com/Vla8islav/metrics-aggregator/internal/repository"
	"go.uber.org/zap"
)

func initRepository(currentConfig *config.Options, logger *zap.Logger) (domain.MetricRepository, error) {
	if currentConfig.DatabaseDSN.BeenSet {
		return repository.NewPostgresStorage(currentConfig, currentConfig.MigrationsFolder.Value)
	}

	dbInMemory := repository.NewMemStorage(currentConfig)
	if !currentConfig.Restore.BeenSet || !currentConfig.Restore.Value {
		logger.Info("making a fresh in-memory DB because the restore flag is false")
		return dbInMemory, nil
	}

	logger.Info("trying to load data from file into the in-memory DB")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go dbInMemory.RunSaver(ctx)

	if err := dbInMemory.Restore(ctx); err != nil {
		return nil, err
	}

	return dbInMemory, nil
}
