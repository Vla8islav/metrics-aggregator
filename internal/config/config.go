package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress  string                `env:"ADDRESS"`
	PollInterval   CustomSecondsDuration `env:"POLL_INTERVAL"`
	ReportInterval CustomSecondsDuration `env:"REPORT_INTERVAL"`

	StoreInterval   CustomSecondsDuration `env:"STORE_INTERVAL"`
	FileStoragePath string                `env:"FILE_STORAGE_PATH"`
	Restore         bool                  `env:"RESTORE"`
}

var optionsInstance *Options

func ReadFlags() *Options {
	if optionsInstance == nil {
		cmdOptions := getCmdOptions()
		envOptions := getEnvOptions()

		finalOptions := Options{}
		// env options are the priority
		mergeOptions(&finalOptions, envOptions)
		mergeOptions(&finalOptions, cmdOptions)
		optionsInstance = &finalOptions
	}
	return optionsInstance
}

func mergeOptions(mergeInto *Options, newValues Options) {
	// TODO: should rewrite it using reflect, probably
	if mergeInto.ServerAddress == "" && newValues.ServerAddress != "" {
		mergeInto.ServerAddress = newValues.ServerAddress
	}

	if mergeInto.PollInterval.Duration == 0 && newValues.PollInterval.Duration != 0 {
		mergeInto.PollInterval = newValues.PollInterval
	}

	if mergeInto.ReportInterval.Duration == 0 && newValues.ReportInterval.Duration != 0 {
		mergeInto.ReportInterval = newValues.ReportInterval
	}

	if mergeInto.StoreInterval.Duration == 0 && newValues.StoreInterval.Duration != 0 {
		mergeInto.StoreInterval = newValues.StoreInterval
	}

	if mergeInto.FileStoragePath == "" && newValues.FileStoragePath != "" {
		mergeInto.FileStoragePath = newValues.FileStoragePath
	}

	if mergeInto.Restore && newValues.ReportInterval.Duration != 0 {
		mergeInto.ReportInterval = newValues.ReportInterval
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

func getCmdOptions() Options {
	opt := Options{
		ReportInterval: CustomSecondsDuration{10 * time.Second},
		PollInterval:   CustomSecondsDuration{2 * time.Second},
	}
	flag.StringVar(&opt.ServerAddress, "a", "localhost:8080", "port on which the server should run")
	flag.Var(&opt.ReportInterval, "r", "how often console utility should send metrics")
	flag.Var(&opt.PollInterval, "p", "how often console utility should poll metrics")

	flag.Var(&opt.StoreInterval, "i", "интервал времени в секундах, по истечении которого"+
		" текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)")
	flag.StringVar(&opt.FileStoragePath, "f", "storage.dat", "путь до файла, куда "+
		"сохраняются текущие значения. Имя файла для значения по умолчанию придумайте сами.")
	flag.BoolVar(&opt.Restore, "r", false, "булево значение (true/false), определяющее, "+
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")

	flag.Parse()
	return opt
}
