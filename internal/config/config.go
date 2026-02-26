package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress  string        `env:"SERVER_ADDRESS"`
	PollInterval   time.Duration `env:"POLL_INTERVAL"`
	ReportInterval time.Duration `env:"REPORT_INTERVAL"`
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

	if mergeInto.PollInterval == 0 && newValues.PollInterval != 0 {
		mergeInto.PollInterval = newValues.PollInterval
	}

	if mergeInto.ReportInterval == 0 && newValues.ReportInterval != 0 {
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
	opt := Options{}
	flag.StringVar(&opt.ServerAddress, "a", "localhost:8080", "port on which the server should run")
	flag.DurationVar(&opt.ReportInterval, "r", 10*time.Second, "how often console utility should send metrics")
	flag.DurationVar(&opt.PollInterval, "p", 2*time.Second, "how often console utility should poll metrics")
	flag.Parse()
	return opt
}
