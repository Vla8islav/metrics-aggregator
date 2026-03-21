package config

import (
	"flag"
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
}

func ReadFlags() *Options {
	var optionsInstance *Options
	cmdOptions := getCmdOptions()
	envOptions := getEnvOptions()

	finalOptions := Options{
		ServerAddress:   OptionalString{Value: "localhost:8080", BeenSet: false},
		PollInterval:    CustomSecondsDuration{Duration: time.Second * 2, BeenSet: false},
		ReportInterval:  CustomSecondsDuration{Duration: time.Second * 10, BeenSet: false},
		StoreInterval:   CustomSecondsDuration{Duration: time.Second * 300, BeenSet: false},
		FileStoragePath: OptionalString{Value: "storage.dat", BeenSet: false},
		Restore:         OptionalBool{Value: true, BeenSet: false},
	}

	// env options are the priority
	mergeOptions(&finalOptions, envOptions)
	mergeOptions(&finalOptions, cmdOptions)
	optionsInstance = &finalOptions
	return optionsInstance
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
		StoreInterval: CustomSecondsDuration{Duration: 300 * time.Second},
	}
	flag.Var(&opt.ServerAddress, "a", "port on which the server should run")
	flag.Var(&opt.ReportInterval, "e", "how often console utility should send metrics")
	flag.Var(&opt.PollInterval, "p", "how often console utility should poll metrics")

	flag.Var(&opt.StoreInterval, "i", "интервал времени в секундах, по истечении которого"+
		" текущие показания сервера сохраняются на диск (по умолчанию 300 секунд, значение 0 делает запись синхронной)")
	flag.Var(&opt.FileStoragePath, "f", "путь до файла, куда "+
		"сохраняются текущие значения. Имя файла для значения по умолчанию придумайте сами.")
	flag.Var(&opt.Restore, "r", "булево значение (true/false), определяющее, "+
		"следует ли загружать ранее сохранённые значения из указанного файла при старте сервера")

	flag.Parse()
	return opt
}
