package main

import (
	"fmt"
	"reflect"

	"github.com/urfave/cli"
	"go.uber.org/zap"
)

func NewDefaultConfig() *Config {
	return &Config{
		Host:         "0.0.0.0",
		HTTPPort:     "8080",
		GRPCPort:     "9090",
		GenesisFile:  "",
		MnemonicFile: "",
		PrivValFile:  "",
		StatusURL:    "http://0.0.0.0:26657/status",
	}
}

type Config struct {
	// Host is the interface to bind the HTTP service on
	Host string `name:"host" json:"host" env:"HOST" usage:"Host address to listen on"`
	// HTTPPort is the port for the http server
	HTTPPort string `name:"http-port" json:"http_port" env:"HTTP_PORT" usage:"Port for http server"`
	// GRPCPort is the port for the grpc server
	GRPCPort string `name:"grpc-port" json:"grpc_port" env:"GRPC_PORT" usage:"Port for gRPC server"`
	// GenesisFile is full path to the genesis file
	GenesisFile string `name:"genesis-file" json:"genesis_file" env:"GENESIS_FILE" usage:"Path of genesis file"`
	// MnemonicFile is full path to the keys file
	MnemonicFile string `name:"mnemonic-file" json:"mnemonic_file" env:"MNEMONIC_FILE" usage:"Path of mnemonic file"`
	// PrivValFile is full path of the node validator private key file
	PrivValFile string `name:"priv-val-file" json:"priv_val_file" env:"PRIV_VAL_FILE" usage:"Path of priv_validator_key.json file for node"`
	// StatusURL is used to fetch status info from blockchain node
	StatusURL string `name:"status-url" json:"status_url" env:"STATUS_URL" usage:"URL to fetch chain status"`
	// Verbose switches on debug logging
	Verbose bool `name:"verbose" json:"verbose" usage:"switch on debug / verbose logging"`
	// OnlyFatalLog set log level as fatal to ignore logs
	OnlyFatalLog bool `name:"only-fatal-log" json:"only-fatal-log" usage:"used while running test"`
}

func GetCommandLineOptions() []cli.Flag {
	defaults := NewDefaultConfig()
	var flags []cli.Flag
	count := reflect.TypeOf(Config{}).NumField()
	for i := 0; i < count; i++ {
		field := reflect.TypeOf(Config{}).Field(i)
		usage, found := field.Tag.Lookup("usage")
		if !found {
			continue
		}
		envName := field.Tag.Get("env")
		if envName != "" {
			envName = envPrefix + envName
		}
		optName := field.Tag.Get("name")

		switch t := field.Type; t.Kind() {
		case reflect.Bool:
			dv := reflect.ValueOf(defaults).Elem().FieldByName(field.Name).Bool()
			msg := fmt.Sprintf("%s (default: %t)", usage, dv)
			flags = append(flags, cli.BoolTFlag{
				Name:   optName,
				Usage:  msg,
				EnvVar: envName,
			})
		case reflect.String:
			defaultValue := reflect.ValueOf(defaults).Elem().FieldByName(field.Name).String()
			flags = append(flags, cli.StringFlag{
				Name:   optName,
				Usage:  usage,
				EnvVar: envName,
				Value:  defaultValue,
			})
		}
	}

	return flags
}

func ParseCLIOptions(cx *cli.Context, config *Config) (err error) {
	// iterate the Config and grab command line options via reflection
	count := reflect.TypeOf(config).Elem().NumField()
	for i := 0; i < count; i++ {
		field := reflect.TypeOf(config).Elem().Field(i)
		name := field.Tag.Get("name")

		if cx.IsSet(name) {
			switch field.Type.Kind() {
			case reflect.Bool:
				reflect.ValueOf(config).Elem().FieldByName(field.Name).SetBool(cx.Bool(name))
			case reflect.String:
				reflect.ValueOf(config).Elem().FieldByName(field.Name).SetString(cx.String(name))
			}
		}
	}
	return nil
}

func NewLogger(config *Config) (*zap.Logger, error) {
	c := zap.NewProductionConfig()
	c.DisableCaller = true
	// c.Encoding = "console"

	if config.Verbose {
		c.DisableCaller = false
		c.Development = true
		c.DisableStacktrace = true // Disable stack trace for development
		c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	if config.OnlyFatalLog {
		c.Level = zap.NewAtomicLevelAt(zap.FatalLevel)
	}

	log, err := c.Build()
	if err != nil {
		return nil, err
	}
	zap.ReplaceGlobals(log) // Set zap global logger
	return log, err
}
