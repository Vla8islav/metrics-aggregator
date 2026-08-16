// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/Vla8islav/metrics-aggregator/internal/helpers"
	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsServer configuration parameters for the metrics server
//
// Values' order of precedence: environment vars, command-line flags, defaults
type OptionsServer struct {
	ServerAddress     OptionalString `env:"ADDRESS" json:"address"`
	ServerAddressGRPC OptionalString `env:"ADDRESS_GRPC" json:"address_grpc"`

	StoreInterval   OptionalSecondsDuration `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath OptionalString          `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore         OptionalBool            `env:"RESTORE" json:"restore"`

	DatabaseDSN      OptionalString `env:"DATABASE_DSN" json:"database_dsn"`
	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER" json:"migrations_folder"`

	AuditURL  OptionalString `env:"AUDIT_URL" json:"audit_url"`
	AuditFile OptionalString `env:"AUDIT_FILE" json:"audit_file"`

	PublicKey  OptionalString `env:"PUBLIC_KEY" json:"public_key"`
	PrivateKey OptionalString `env:"PRIVATE_KEY" json:"private_key"`

	// CIDR subnets, comma-separated like 192.168.0.0/24,10.0.0.0/8
	TrustedSubnets OptionalString `env:"TRUSTED_SUBNETS" json:"trusted_subnets"`

	Config OptionalString `env:"CONFIG" json:"-"`

	logger *zap.Logger
}

func logSetFlagsServer(options *OptionsServer) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("-a", options.ServerAddress.Value))
	}

	if options.StoreInterval.BeenSet {
		fields = append(fields, zap.String("-i", options.StoreInterval.Duration.String()))
	}

	if options.FileStoragePath.BeenSet {
		fields = append(fields, zap.String("-f", options.FileStoragePath.Value))
	}

	if options.Restore.BeenSet {
		fields = append(fields, zap.Bool("-r", options.Restore.Value))
	}

	if options.DatabaseDSN.BeenSet {
		fields = append(fields, zap.String("-d", options.DatabaseDSN.Value))
	}

	if options.MigrationsFolder.BeenSet {
		fields = append(fields, zap.String("-m", options.MigrationsFolder.Value))
	}

	if options.AuditURL.BeenSet {
		fields = append(fields, zap.String("-audit-url", options.AuditURL.Value))
	}

	if options.AuditFile.BeenSet {
		fields = append(fields, zap.String("-audit-file", options.AuditFile.Value))
	}

	if options.PublicKey.BeenSet {
		fields = append(fields, zap.String("-public-key", options.PublicKey.Value))
	}

	if options.PrivateKey.BeenSet {
		fields = append(fields, zap.String("-private-key", options.PrivateKey.Value))
	}

	if options.Config.BeenSet {
		fields = append(fields, zap.String("-config", options.Config.Value))
	}

	if options.TrustedSubnets.BeenSet {
		fields = append(fields, zap.String("-t", options.TrustedSubnets.Value))
	}

	if options.ServerAddressGRPC.BeenSet {
		fields = append(fields, zap.String("-address-grpc", options.ServerAddressGRPC.Value))
	}

	if len(fields) == 0 {
		options.logger.Info("no command-line flags were set")
		return
	}

	options.logger.Info("command line options", fields...)
}

func logSetEnvServer(options *OptionsServer) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("ADDRESS", options.ServerAddress.Value))
	}

	if options.StoreInterval.BeenSet {
		fields = append(fields, zap.String("STORE_INTERVAL", options.StoreInterval.Duration.String()))
	}

	if options.FileStoragePath.BeenSet {
		fields = append(fields, zap.String("FILE_STORAGE_PATH", options.FileStoragePath.Value))
	}

	if options.Restore.BeenSet {
		fields = append(fields, zap.Bool("RESTORE", options.Restore.Value))
	}

	if options.DatabaseDSN.BeenSet {
		fields = append(fields, zap.String("DATABASE_DSN", options.DatabaseDSN.Value))
	}

	if options.MigrationsFolder.BeenSet {
		fields = append(fields, zap.String("MIGRATIONS_FOLDER", options.MigrationsFolder.Value))
	}

	if options.AuditURL.BeenSet {
		fields = append(fields, zap.String("AUDIT_URL", options.AuditURL.Value))
	}

	if options.AuditFile.BeenSet {
		fields = append(fields, zap.String("AUDIT_FILE", options.AuditFile.Value))
	}

	if options.PublicKey.BeenSet {
		fields = append(fields, zap.String("PUBLIC_KEY", options.PublicKey.Value))
	}

	if options.PrivateKey.BeenSet {
		fields = append(fields, zap.String("PRIVATE_KEY", options.PrivateKey.Value))
	}

	if options.Config.BeenSet {
		fields = append(fields, zap.String("CONFIG", options.Config.Value))
	}

	if options.TrustedSubnets.BeenSet {
		fields = append(fields, zap.String("TRUSTED_SUBNETS", options.TrustedSubnets.Value))
	}

	if options.ServerAddressGRPC.BeenSet {
		fields = append(fields, zap.String("ADDRESS_GRPC", options.ServerAddressGRPC.Value))
	}

	if len(fields) == 0 {
		options.logger.Info("no environment variables were set")
		return
	}

	options.logger.Info("environment variables", fields...)
}

func logConfigOptionsServer(options *OptionsServer) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("address", options.ServerAddress.Value))
	}

	if options.StoreInterval.BeenSet {
		fields = append(fields, zap.String("store_interval", options.StoreInterval.Duration.String()))
	}

	if options.FileStoragePath.BeenSet {
		fields = append(fields, zap.String("store_file", options.FileStoragePath.Value))
	}

	if options.Restore.BeenSet {
		fields = append(fields, zap.Bool("restore", options.Restore.Value))
	}

	if options.DatabaseDSN.BeenSet {
		fields = append(fields, zap.String("database_dsn", options.DatabaseDSN.Value))
	}

	if options.MigrationsFolder.BeenSet {
		fields = append(fields, zap.String("migrations_folder", options.MigrationsFolder.Value))
	}

	if options.AuditURL.BeenSet {
		fields = append(fields, zap.String("audit_url", options.AuditURL.Value))
	}

	if options.AuditFile.BeenSet {
		fields = append(fields, zap.String("audit_file", options.AuditFile.Value))
	}

	if options.PublicKey.BeenSet {
		fields = append(fields, zap.String("public_key", options.PublicKey.Value))
	}

	if options.PrivateKey.BeenSet {
		fields = append(fields, zap.String("private_key", options.PrivateKey.Value))
	}

	if options.TrustedSubnets.BeenSet {
		fields = append(fields, zap.String("trusted_subnets", options.TrustedSubnets.Value))
	}

	if options.ServerAddressGRPC.BeenSet {
		fields = append(fields, zap.String("address_grpc", options.ServerAddressGRPC.Value))
	}

	if len(fields) == 0 {
		options.logger.Info("no config file options were set")
		return
	}

	options.logger.Info("config file options", fields...)
}

// ReadFlagsServer reads server configuration from command-line arguments and environment variables.
//
// precedence: environment variables, command-line flags, config file, defaults.
func ReadFlagsServer(args []string, logger *zap.Logger) (*OptionsServer, error) {
	if logger == nil {
		panic("config server logger is nil")
	}

	cmdOptions, err := getOptionsServer(args, logger)
	if err != nil {
		logger.Fatal("failed to read command-line flags", zap.Error(err))
	}
	logSetFlagsServer(cmdOptions)

	envOptions, err := getEnvOptionsServer(logger)
	if err != nil {
		logger.Fatal("failed to read environment variables", zap.Error(err))
	}
	logSetEnvServer(envOptions)

	var diskConfigOptions OptionsServer
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		// We need to read the config file before assembling the full consensus.
		var configFilename string

		if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		} else if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		}

		diskConfigOptions, err = getDiskConfigOptionsServer(configFilename, logger)
		if err != nil {
			logger.Fatal("failed to read config file", zap.Error(err))
		}

		logConfigOptionsServer(&diskConfigOptions)
	}

	finalOptions := OptionsServer{
		ServerAddress:     OptionalString{Value: "localhost:8080", BeenSet: false},
		ServerAddressGRPC: OptionalString{Value: "localhost:9090", BeenSet: false},
		StoreInterval:     OptionalSecondsDuration{Duration: time.Second * 300, BeenSet: false},
		FileStoragePath:   OptionalString{Value: "storage.dat", BeenSet: false},
		Restore:           OptionalBool{Value: true, BeenSet: false},
		DatabaseDSN: OptionalString{
			Value:   "postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable",
			BeenSet: false,
		},
		MigrationsFolder: OptionalString{Value: "./migrations", BeenSet: false},
		AuditFile:        OptionalString{Value: "", BeenSet: false},
		AuditURL:         OptionalString{Value: "", BeenSet: false},
		PublicKey:        OptionalString{Value: "", BeenSet: false},
		PrivateKey:       OptionalString{Value: "", BeenSet: false},
		TrustedSubnets:   OptionalString{Value: "", BeenSet: false},
		Config:           OptionalString{Value: "", BeenSet: false},
		logger:           logger,
	}

	// Environment options have the highest priority,
	// then command-line options, then disk config options.
	mergeOptionsServer(&finalOptions, diskConfigOptions)
	mergeOptionsServer(&finalOptions, *cmdOptions)
	mergeOptionsServer(&finalOptions, *envOptions)

	if err = sanityCheckConfig(&finalOptions); err != nil {
		return nil, err
	}

	return &finalOptions, nil
}

func sanityCheckConfig(optionsSrv *OptionsServer) error {
	if optionsSrv == nil {
		return fmt.Errorf("optionsSrv is nil")
	}

	if !optionsSrv.TrustedSubnets.BeenSet {
		return nil
	}

	for _, rawSubnet := range strings.Split(optionsSrv.TrustedSubnets.Value, ",") {
		subnet := strings.TrimSpace(rawSubnet)
		if err := helpers.ValidateMaskCIDR(subnet); err != nil {
			return fmt.Errorf("invalid subnet %q: %w", subnet, err)
		}
	}

	return nil
}

func getDiskConfigOptionsServer(filename string, logger *zap.Logger) (OptionsServer, error) {
	if filename == "" {
		return OptionsServer{logger: logger}, nil
	}

	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsServer{logger: logger}, err
	}

	options := OptionsServer{
		logger: logger,
	}

	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsServer{logger: logger}, err
	}

	return options, nil
}

func mergeOptionsServer(mergeInto *OptionsServer, newValues OptionsServer) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
		mergeInto.ServerAddress.BeenSet = true
	}

	if newValues.ServerAddressGRPC.BeenSet {
		mergeInto.ServerAddressGRPC = newValues.ServerAddressGRPC
		mergeInto.ServerAddressGRPC.BeenSet = true
	}

	if newValues.StoreInterval.BeenSet {
		mergeInto.StoreInterval = newValues.StoreInterval
		mergeInto.StoreInterval.BeenSet = true
	}

	if newValues.FileStoragePath.BeenSet {
		mergeInto.FileStoragePath = newValues.FileStoragePath
		mergeInto.FileStoragePath.BeenSet = true
	}

	if newValues.Restore.BeenSet {
		mergeInto.Restore = newValues.Restore
		mergeInto.Restore.BeenSet = true
	}

	if newValues.DatabaseDSN.BeenSet {
		mergeInto.DatabaseDSN = newValues.DatabaseDSN
		mergeInto.DatabaseDSN.BeenSet = true
	}

	if newValues.MigrationsFolder.BeenSet {
		mergeInto.MigrationsFolder = newValues.MigrationsFolder
		mergeInto.MigrationsFolder.BeenSet = true
	}

	if newValues.AuditURL.BeenSet {
		mergeInto.AuditURL = newValues.AuditURL
		mergeInto.AuditURL.BeenSet = true
	}

	if newValues.AuditFile.BeenSet {
		mergeInto.AuditFile = newValues.AuditFile
		mergeInto.AuditFile.BeenSet = true
	}

	if newValues.PublicKey.BeenSet {
		mergeInto.PublicKey = newValues.PublicKey
		mergeInto.PublicKey.BeenSet = true
	}

	if newValues.PrivateKey.BeenSet {
		mergeInto.PrivateKey = newValues.PrivateKey
		mergeInto.PrivateKey.BeenSet = true
	}

	if newValues.TrustedSubnets.BeenSet {
		mergeInto.TrustedSubnets = newValues.TrustedSubnets
		mergeInto.TrustedSubnets.BeenSet = true
	}

	if newValues.Config.BeenSet {
		mergeInto.Config = newValues.Config
		mergeInto.Config.BeenSet = true
	}
}

func getEnvOptionsServer(logger *zap.Logger) (*OptionsServer, error) {
	opt := OptionsServer{
		logger: logger,
	}

	if err := env.Parse(&opt); err != nil {
		return nil, err
	}

	return &opt, nil
}

func getOptionsServer(args []string, logger *zap.Logger) (*OptionsServer, error) {
	opt := &OptionsServer{
		logger: logger,
	}

	fs := flag.NewFlagSet("metrics-aggregator-server", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.Var(&opt.ServerAddress, "a", "port on which the server should run")
	fs.Var(&opt.ServerAddressGRPC, "address-grpc", "port on which the grpc server should run")

	fs.Var(&opt.StoreInterval, "i",
		"интервал времени в секундах, по истечении которого текущие показания сервера сохраняются на диск "+
			"(по умолчанию 300 секунд, значение 0 делает запись синхронной)",
	)
	fs.Var(
		&opt.FileStoragePath, "f", "путь до файла, куда сохраняются текущие значения",
	)
	fs.Var(&opt.Restore, "r", "булево значение (true/false), определяющее, следует ли загружать ранее"+
		" сохранённые значения из указанного файла при старте сервера",
	)

	fs.Var(&opt.DatabaseDSN, "d", "connection string/dsn для postgres базы данных")
	fs.Var(&opt.MigrationsFolder, "m", "относительный путь до миграций, например ./migrations")

	fs.Var(&opt.AuditURL, "audit-url", "адрес сервера аудита")
	fs.Var(&opt.AuditFile, "audit-file", "путь до файла аудита")

	fs.Var(&opt.PublicKey, "public-key", "путь до файла с публичным ключом")
	fs.Var(&opt.PrivateKey, "private-key", "путь до файла с приватным ключом")

	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")

	fs.Var(&opt.TrustedSubnets, "t", "CIDR допустимых подсетей, разделённых запятой")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opt, nil
}
