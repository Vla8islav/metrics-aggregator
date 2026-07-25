// Package config parsing passed config values
package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"
	"time"

	"github.com/caarlos0/env/v6"
	"go.uber.org/zap"
)

// OptionsClient configuration parameters for the metrics agent client
// the order of precedence: env, command line, default value
type OptionsClient struct {
	ServerAddress     OptionalString          `env:"ADDRESS" json:"address"`
	ServerAddressGRPC OptionalString          `env:"ADDRESS_GRPC" json:"address_grpc"`
	PollInterval      OptionalSecondsDuration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval    OptionalSecondsDuration `env:"REPORT_INTERVAL" json:"report_interval"`
	RateLimit         OptionalInt             `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey         OptionalString          `env:"CRYPTO_KEY" json:"crypto_key"`

	SecretKey OptionalString `env:"KEY" json:"secret_key"`

	Config OptionalString `env:"CONFIG" json:"-"`

	logger *zap.Logger
}

func logSetFlagsClient(options *OptionsClient) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("-a", options.ServerAddress.String()))
	}

	if options.ReportInterval.BeenSet {
		fields = append(fields, zap.String("-r", options.ReportInterval.String()))
	}

	if options.PollInterval.BeenSet {
		fields = append(fields, zap.String("-p", options.PollInterval.Duration.String()))
	}

	if options.SecretKey.BeenSet {
		fields = append(fields, zap.String("-k", options.SecretKey.Value))
	}

	if options.RateLimit.BeenSet {
		fields = append(fields, zap.Int("-l", options.RateLimit.Value))
	}

	if options.CryptoKey.BeenSet {
		fields = append(fields, zap.String("-crypto-key", options.CryptoKey.Value))
	}

	if options.Config.BeenSet {
		fields = append(fields, zap.String("-config", options.Config.Value))
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

func logSetEnvClient(options *OptionsClient) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("ADDRESS", options.ServerAddress.Value))
	}

	if options.ReportInterval.BeenSet {
		fields = append(fields, zap.String("REPORT_INTERVAL", options.ReportInterval.Duration.String()))
	}

	if options.PollInterval.BeenSet {
		fields = append(fields, zap.String("POLL_INTERVAL", options.PollInterval.Duration.String()))
	}

	if options.SecretKey.BeenSet {
		fields = append(fields, zap.String("KEY", options.SecretKey.Value))
	}

	if options.RateLimit.BeenSet {
		fields = append(fields, zap.Int("RATE_LIMIT", options.RateLimit.Value))
	}

	if options.CryptoKey.BeenSet {
		fields = append(fields, zap.String("CRYPTO_KEY", options.CryptoKey.Value))
	}

	if options.Config.BeenSet {
		fields = append(fields, zap.String("CONFIG", options.Config.Value))
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

func logConfigOptionsClient(options *OptionsClient) {
	if options == nil {
		return
	}

	fields := make([]zap.Field, 0)

	if options.ServerAddress.BeenSet {
		fields = append(fields, zap.String("address", options.ServerAddress.Value))
	}

	if options.PollInterval.BeenSet {
		fields = append(fields, zap.String("poll_interval", options.PollInterval.Duration.String()))
	}

	if options.ReportInterval.BeenSet {
		fields = append(fields, zap.String("report_interval", options.ReportInterval.Duration.String()))
	}

	if options.RateLimit.BeenSet {
		fields = append(fields, zap.Int("rate_limit", options.RateLimit.Value))
	}

	if options.SecretKey.BeenSet {
		fields = append(fields, zap.String("secret_key", options.SecretKey.Value))
	}

	if options.CryptoKey.BeenSet {
		fields = append(fields, zap.String("crypto_key", options.CryptoKey.Value))
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

// ReadFlagsClient reads and merges client configuration from command-line flags and environment variables
func ReadFlagsClient(args []string, logger *zap.Logger) *OptionsClient {
	if logger == nil {
		panic("config client logger is nil")
	}

	cmdOptions, err := getOptionsClient(args, logger)
	if err != nil {
		logger.Fatal("failed to read command-line flags", zap.Error(err))
	}
	logSetFlagsClient(cmdOptions)

	envOptions, err := getEnvOptionsClient(logger)
	if err != nil {
		logger.Fatal("failed to read environment variables", zap.Error(err))
	}
	logSetEnvClient(envOptions)

	var diskConfigOptions OptionsClient
	if cmdOptions.Config.BeenSet || envOptions.Config.BeenSet {
		// we need to read the config file before assembling the full consensus
		var configFilename string
		if cmdOptions.Config.BeenSet && cmdOptions.Config.Value != "" {
			configFilename = cmdOptions.Config.Value
		} else if envOptions.Config.BeenSet && envOptions.Config.Value != "" {
			configFilename = envOptions.Config.Value
		}
		diskConfigOptions, err = getDiskConfigOptionsClient(configFilename, logger)
		if err != nil {
			logger.Fatal("failed to read config file", zap.Error(err))
		}
		logConfigOptionsClient(&diskConfigOptions)
	}

	finalOptions := OptionsClient{
		ServerAddress:     OptionalString{Value: "localhost:8080", BeenSet: false},
		ServerAddressGRPC: OptionalString{Value: "localhost:9090", BeenSet: false},
		PollInterval:      OptionalSecondsDuration{Duration: time.Second * 2, BeenSet: false},
		ReportInterval:    OptionalSecondsDuration{Duration: time.Second * 10, BeenSet: false},
		SecretKey:         OptionalString{Value: "", BeenSet: false},
		RateLimit:         OptionalInt{Value: 10, BeenSet: false},
		CryptoKey:         OptionalString{Value: "", BeenSet: false},
		Config:            OptionalString{Value: "", BeenSet: false},
		logger:            logger,
	}

	// env options are the priority, then cmd options, then disk options
	mergeOptionsClient(&finalOptions, diskConfigOptions)
	mergeOptionsClient(&finalOptions, *cmdOptions)
	mergeOptionsClient(&finalOptions, *envOptions)

	//setOptionsTrue(&finalOptions)
	return &finalOptions
}

func getDiskConfigOptionsClient(filename string, logger *zap.Logger) (OptionsClient, error) {
	if filename == "" {
		return OptionsClient{logger: logger}, nil
	}

	configBytes, err := os.ReadFile(filename)
	if err != nil {
		return OptionsClient{logger: logger}, err
	}

	options := OptionsClient{logger: logger}
	if err = json.Unmarshal(configBytes, &options); err != nil {
		return OptionsClient{logger: logger}, err
	}

	return options, nil
}

func mergeOptionsClient(mergeInto *OptionsClient, newValues OptionsClient) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
		mergeInto.ServerAddress.BeenSet = true
	}

	if newValues.PollInterval.BeenSet {
		mergeInto.PollInterval = newValues.PollInterval
		mergeInto.PollInterval.BeenSet = true
	}

	if newValues.ReportInterval.BeenSet {
		mergeInto.ReportInterval = newValues.ReportInterval
		mergeInto.ReportInterval.BeenSet = true
	}

	if newValues.SecretKey.BeenSet {
		mergeInto.SecretKey = newValues.SecretKey
		mergeInto.SecretKey.BeenSet = true
	}

	if newValues.RateLimit.BeenSet {
		mergeInto.RateLimit = newValues.RateLimit
		mergeInto.RateLimit.BeenSet = true
	}

	if newValues.CryptoKey.BeenSet {
		mergeInto.CryptoKey = newValues.CryptoKey
		mergeInto.CryptoKey.BeenSet = true
	}

	if newValues.Config.BeenSet {
		mergeInto.Config = newValues.Config
		mergeInto.Config.BeenSet = true
	}

	if newValues.ServerAddressGRPC.BeenSet {
		mergeInto.ServerAddressGRPC = newValues.ServerAddressGRPC
		mergeInto.ServerAddressGRPC.BeenSet = true
	}
}

func getEnvOptionsClient(logger *zap.Logger) (*OptionsClient, error) {
	opt := OptionsClient{logger: logger}
	err := env.Parse(&opt)
	if err != nil {
		return nil, err
	}
	return &opt, nil
}

func getOptionsClient(args []string, logger *zap.Logger) (*OptionsClient, error) {

	opt := &OptionsClient{logger: logger}

	fs := flag.NewFlagSet("metrics-aggregator-client", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // optional: silence flag errors in tests

	fs.Var(&opt.ServerAddress, "a", "port on which the http server should run")
	fs.Var(&opt.ServerAddressGRPC, "address-grpc", "port on which the grpc server should run")

	fs.Var(&opt.ReportInterval, "r", "how often console utility should send metrics")
	fs.Var(&opt.PollInterval, "p", "how often console utility should poll metrics")
	fs.Var(&opt.SecretKey, "k", "симметричный ключ шифрования для подписи сообщений")
	fs.Var(&opt.RateLimit, "l", "потолок одновременных запросов делается к серверу, RATE_LIMIT")

	fs.Var(&opt.CryptoKey, "crypto-key", "путь до файла с публичным ключом")

	fs.Var(&opt.Config, "config", "путь до файла с конфигурацией приложения")
	fs.Var(&opt.Config, "c", "путь до файла с конфигурацией приложения")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opt, nil
}
