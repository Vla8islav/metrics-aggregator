package config

import (
	"flag"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

type Options struct {
	ServerAddress  string                `env:"SERVER_ADDRESS"`
	PollInterval   CustomSecondsDuration `env:"POLL_INTERVAL"`
	ReportInterval CustomSecondsDuration `env:"REPORT_INTERVAL"`
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
	flag.Parse()
	return opt
}
