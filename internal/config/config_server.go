package config

import (
	"flag"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/caarlos0/env/v6"
)

// OptionsServer TODO: implement a clean option separation
type OptionsServer struct {
	ServerAddress  OptionalString        `env:"ADDRESS"`
	PollInterval   CustomSecondsDuration `env:"POLL_INTERVAL"`
	ReportInterval CustomSecondsDuration `env:"REPORT_INTERVAL"`

	StoreInterval   CustomSecondsDuration `env:"STORE_INTERVAL"`
	FileStoragePath OptionalString        `env:"FILE_STORAGE_PATH"`
	Restore         OptionalBool          `env:"RESTORE"`

	DatabaseDSN      OptionalString `env:"DATABASE_DSN"`
	MigrationsFolder OptionalString `env:"MIGRATIONS_FOLDER"`

	SecretKey OptionalString `env:"SECRET_KEY"`
}

func logSetFlagsServer(options *OptionsServer) {
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

	if options.StoreInterval.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-i=%s", options.StoreInterval.Duration))
	}

	if options.FileStoragePath.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-f=%s", options.FileStoragePath.Value))
	}

	if options.Restore.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-e=%t", options.Restore.Value))
	}

	if options.DatabaseDSN.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-d=%s", options.DatabaseDSN.Value))
	}

	if options.MigrationsFolder.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-m=%s", options.MigrationsFolder.Value))
	}

	if options.SecretKey.BeenSet {
		setFlags = append(setFlags, fmt.Sprintf("-k=%s", options.SecretKey.Value))
	}

	if len(setFlags) == 0 {
		log.Println("no command-line flags were set")
		return
	}

	for _, flagValue := range setFlags {
		log.Printf("command-line flag set: %s", flagValue)
	}
}

func logSetEnvServer(options *OptionsServer) {
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

	if options.StoreInterval.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("STORE_INTERVAL=%s", options.StoreInterval.Duration))
	}

	if options.FileStoragePath.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("FILE_STORAGE_PATH=%s", options.FileStoragePath.Value))
	}

	if options.Restore.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("RESTORE=%t", options.Restore.Value))
	}

	if options.DatabaseDSN.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("DATABASE_DSN=%s", options.DatabaseDSN.Value))
	}

	if options.MigrationsFolder.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("MIGRATIONS_FOLDER=%s", options.MigrationsFolder.Value))
	}

	if options.SecretKey.BeenSet {
		setEnv = append(setEnv, fmt.Sprintf("SECRET_KEY=%s", options.SecretKey.Value))
	}

	if len(setEnv) == 0 {
		log.Println("no environment variables were set")
		return
	}

	for _, envValue := range setEnv {
		log.Printf("environment variable set: %s", envValue)
	}
}

func ReadFlags(args []string) *OptionsServer {
	cmdOptions, err := getOptionsServer(args)
	if err != nil {
		log.Fatalln(err)
	}
	logSetFlagsServer(cmdOptions)

	envOptions := getEnvOptions()
	logSetEnvServer(envOptions)

	finalOptions := OptionsServer{
		ServerAddress:   OptionalString{Value: "localhost:8080", BeenSet: false},
		PollInterval:    CustomSecondsDuration{Duration: time.Second * 2, BeenSet: false},
		ReportInterval:  CustomSecondsDuration{Duration: time.Second * 10, BeenSet: false},
		StoreInterval:   CustomSecondsDuration{Duration: time.Second * 300, BeenSet: false},
		FileStoragePath: OptionalString{Value: "storage.dat", BeenSet: false},
		DatabaseDSN: OptionalString{Value: "postgres://default_user:default_password@localhost:5432/metrics_db?sslmode=disable",
			BeenSet: false},
		MigrationsFolder: OptionalString{Value: "./migrations", BeenSet: false},
		Restore:          OptionalBool{Value: true, BeenSet: false},
		SecretKey:        OptionalString{Value: "", BeenSet: false},
	}

	// env options are the priority
	mergeOptionsServer(&finalOptions, *cmdOptions)
	mergeOptionsServer(&finalOptions, *envOptions)

	//setOptionsTrue(&finalOptions)
	return &finalOptions
}

func mergeOptionsServer(mergeInto *OptionsServer, newValues OptionsServer) {
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

	if newValues.SecretKey.BeenSet {
		mergeInto.SecretKey = newValues.SecretKey
		mergeInto.SecretKey.BeenSet = true
	}
}

func getEnvOptions() *OptionsServer {
	var opt OptionsServer
	err := env.Parse(&opt)
	if err != nil {
		log.Fatalln(err)
	}
	return &opt
}

func getOptionsServer(args []string) (*OptionsServer, error) {

	opt := &OptionsServer{}

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
	fs.Var(&opt.MigrationsFolder, "m", "относительный путь до миграций, например ./migrations")
	fs.Var(&opt.SecretKey, "k", "симметричный ключ шифрования для подписи сообщений")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	return opt, nil
}
