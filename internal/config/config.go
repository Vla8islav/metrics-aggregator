package config

import (
	"flag"
	"io"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress  OptionalString        `env:"ADDRESS"`
	PollInterval   CustomSecondsDuration `env:"POLL_INTERVAL"`
	ReportInterval CustomSecondsDuration `env:"REPORT_INTERVAL"`

	StoreInterval   CustomSecondsDuration `env:"STORE_INTERVAL"`
	FileStoragePath OptionalString        `env:"FILE_STORAGE_PATH"`
	Restore         OptionalBool          `env:"RESTORE"`

	DatabaseDSN      OptionalString `env:"DATABASE_DSN"`
	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER"`
}

func setOptionsTrue(options *Options) {
	options.ServerAddress.BeenSet = true
	options.PollInterval.BeenSet = true
	options.ReportInterval.BeenSet = true
	options.StoreInterval.BeenSet = true
	options.FileStoragePath.BeenSet = true
	options.Restore.BeenSet = true
	options.DatabaseDSN.BeenSet = true
	options.MigrationsFolder.BeenSet = true

}

func ReadFlags(args []string) *Options {
	cmdOptions, err := getServerOptions(args)
	if err != nil {
		log.Fatalln(err)
	}

	envOptions := getEnvOptions()

	finalOptions := Options{
		ServerAddress:    OptionalString{Value: "localhost:8080", BeenSet: false},
		PollInterval:     CustomSecondsDuration{Duration: time.Second * 2, BeenSet: false},
		ReportInterval:   CustomSecondsDuration{Duration: time.Second * 10, BeenSet: false},
		StoreInterval:    CustomSecondsDuration{Duration: time.Second * 300, BeenSet: false},
		FileStoragePath:  OptionalString{Value: "storage.dat", BeenSet: false},
		DatabaseDSN:      OptionalString{Value: "postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable", BeenSet: false},
		MigrationsFolder: OptionalString{Value: "./migrations", BeenSet: false},
		Restore:          OptionalBool{Value: true, BeenSet: false},
	}

	// env options are the priority
	mergeOptions(&finalOptions, cmdOptions)
	mergeOptions(&finalOptions, envOptions)

	setOptionsTrue(&finalOptions)
	return &finalOptions
}

func mergeOptions(mergeInto *Options, newValues Options) {
	if newValues.ServerAddress.BeenSet {
		mergeInto.ServerAddress = newValues.ServerAddress
	}

	if newValues.PollInterval.BeenSet {
		mergeInto.PollInterval = newValues.PollInterval
	}

	if newValues.ReportInterval.BeenSet {
		mergeInto.ReportInterval = newValues.ReportInterval
	}

	if newValues.StoreInterval.BeenSet {
		mergeInto.StoreInterval = newValues.StoreInterval
	}

	if newValues.FileStoragePath.BeenSet {
		mergeInto.FileStoragePath = newValues.FileStoragePath
	}

	if newValues.Restore.BeenSet {
		mergeInto.Restore = newValues.Restore
	}

	if newValues.DatabaseDSN.BeenSet {
		mergeInto.DatabaseDSN = newValues.DatabaseDSN
	}

	if newValues.MigrationsFolder.BeenSet {
		mergeInto.MigrationsFolder = newValues.MigrationsFolder
	}
}

func getEnvOptions() Options {
	var opt Options
	err := env.Parse(&opt)
	if err != nil {
		log.Fatalln(err)
	}
	return opt
}

func getServerOptions(args []string) (Options, error) {

	opt := Options{}

	fs := flag.NewFlagSet("metrics-aggregator", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // optional: silence flag errors in tests

	fs.Var(&opt.ServerAddress, "a", "port on which the server should run")
	fs.Var(&opt.ReportInterval, "r", "how often console utility should send metrics")
	fs.Var(&opt.PollInterval, "p", "how often console utility should poll metrics")

	fs.Var(&opt.StoreInterval, "i", "интервал времени в секундах, по истечении которого"+
		" текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)")
	fs.Var(&opt.FileStoragePath, "f", "путь до файла, куда "+
		"сохраняются текущие значения. Имя файла для значения по умолчанию придумайте сами.")
	fs.Var(&opt.Restore, "e", "булево значение (true/false), определяющее, "+
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")
	fs.Var(&opt.DatabaseDSN, "d", "connection string/dsn для postgres базы данных")
	fs.Var(&opt.DatabaseDSN, "m", "относительный путь до миграций, например ./migrations")

	if err := fs.Parse(args); err != nil {
		return Options{}, err
	}

	return opt, nil
}
