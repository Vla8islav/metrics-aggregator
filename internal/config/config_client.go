// Package config parsing passed config values
package config

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

// OptionsClient configuration parameters for the metrics agent client
// the order of precedence: env, command line, default value
type OptionsClient struct {
	ServerAddress  OptionalString          `env:"ADDRESS" json:"address"`
	PollInterval   OptionalSecondsDuration `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval OptionalSecondsDuration `env:"REPORT_INTERVAL" json:"report_interval"`
	RateLimit      OptionalInt             `env:"RATE_LIMIT" json:"rate_limit"`
	CryptoKey      OptionalString          `env:"CRYPTO_KEY" json:"crypto_key"`

	SecretKey OptionalString `env:"KEY" json:"secret_key"`

	Config OptionalString `env:"Config" json:"-"`
}

func logSetFlagsClient(options *OptionsClient) {
	if options == nil {
		return
	}
	var setFlags []string

	if options.ServerAddress.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-a=%s", options.ServerAddress.Value))
	}

	if options.ReportInterval.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-r=%s", options.ReportInterval.Duration))
	}

	if options.PollInterval.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-p=%s", options.PollInterval.Duration))
	}

	if options.SecretKey.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-k=%s", options.SecretKey.Value))
	}

	if options.RateLimit.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-l=%d", options.RateLimit.Value))
	}

	if options.CryptoKey.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-crypto-key=%d", options.CryptoKey.Value))
	}

	if options.Config.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-config=%d", options.Config.Value))
	}

	if len(setFlags) == 0 {
		log.Println("no command-line flags were set")
		return
	}

	for _, flagValue := range setFlags {
		log.Printf("command-line flag set: %s", flagValue)
	}
}

func logSetEnvClient(options *OptionsClient) {
	if options == nil {
		return
	}
	var setEnv []string

	if options.ServerAddress.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("ADDRESS=%s", options.ServerAddress.Value))
	}

	if options.ReportInterval.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("REPORT_INTERVAL=%s", options.ReportInterval.Duration))
	}

	if options.PollInterval.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("POLL_INTERVAL=%s", options.PollInterval.Duration))
	}

	if options.SecretKey.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("SECRET_KEY=%s", options.SecretKey.Value))
	}

	if options.RateLimit.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("RATE_LIMIT=%d", options.RateLimit.Value))
	}

	if options.CryptoKey.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("CRYPTO_KEY=%d", options.CryptoKey.Value))
	}

	if options.Config.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("CONFIG=%d", options.Config.Value))
	}

	if len(setEnv) == 0 {
		log.Println("no environment variables were set")
		return
	}

	for _, envValue := range setEnv {
		log.Printf("environment variable set: %s", envValue)
	}
}

// ReadFlagsClient reads and merges client configuration from command-line flags and environment variables
func ReadFlagsClient(args []string) *OptionsClient {
	cmdOptions, err := getOptionsClient(args)
	if err != nil {
		log.Fatalln(err)
	}
	logSetFlagsClient(cmdOptions)

	envOptions := getEnvOptionsClient()
	logSetEnvClient(envOptions)

	finalOptions := OptionsClient{
		ServerAddress:  OptionalString{Value: "localhost:8080", BeenSet: false},
		PollInterval:   OptionalSecondsDuration{Duration: time.Second * 2, BeenSet: false},
		ReportInterval: OptionalSecondsDuration{Duration: time.Second * 10, BeenSet: false},
		SecretKey:      OptionalString{Value: "", BeenSet: false},
		RateLimit:      OptionalInt{Value: 10, BeenSet: false},
		CryptoKey:      OptionalString{Value: "", BeenSet: false},
		Config:         OptionalString{Value: "", BeenSet: false},
	}

	// env options are the priority
	mergeOptionsClient(&finalOptions, *cmdOptions)
	mergeOptionsClient(&finalOptions, *envOptions)

	//setOptionsTrue(&finalOptions)
	return &finalOptions
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
}

func getEnvOptionsClient() *OptionsClient {
	var opt OptionsClient
	err := env.Parse(&opt)
	if err != nil {
		log.Fatalln(err)
	}
	return &opt
}

func getOptionsClient(args []string) (*OptionsClient, error) {

	opt := &OptionsClient{}

	fs := flag.NewFlagSet("metrics-aggregator-client", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // optional: silence flag errors in tests

	fs.Var(&opt.ServerAddress, "a", "port on which the server should run")
	fs.Var(&opt.ReportInterval, "r", "how often console utility should send metrics")
	fs.Var(&opt.PollInterval, "p", "how often console utility should poll metrics")
	fs.Var(&opt.SecretKey, "k", "симметричный ключ шифрования для подписи сообщений")
	fs.Var(&opt.RateLimit, "l", "потолок одновременных запросов делается к серверу, RATE_LIMIT")

	fs.Var(&opt.CryptoKey, "crypto-key", "путь до файла с публичным ключом")
	fs.Var(&opt.CryptoKey, "config", "путь до файла с конфигурацией приложения")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opt, nil
}
