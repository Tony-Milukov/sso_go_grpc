package config

import (
	"errors"
	"flag"
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

type GrpcConfig struct {
	Port    int    `yaml:"port" env-required`
	Timeout string `yaml:"timeout" env-default:"12h"`
}
type Config struct {
	Env       string `yaml:"env" env-required`
	DbLink    string `yaml:"db_link" env-required`
	DbType    string `yaml:"db_type" env-required`
	JwtSecret string `yaml:"jwt_secret" env-required`
	GRPC      GrpcConfig
}

// MustLoad returns a config by config path which was gotten from getConfigPath
func MustLoad() *Config {
	op := "config.New"
	var cfg Config
	configPath := getConfigPath()

	// check if the config file exists
	if _, err := os.Stat(configPath); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			panic(fmt.Sprintf(`%s: Config dont exist on %s`, op, configPath))
		} else {
			panic(fmt.Sprintf("%s: %w", op, err))
		}
	}

	// read the config and return config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic(fmt.Errorf("%s : %w", op, err))
	}

	return &cfg
}

// getConfigPath this function returns config struct path = flag > os.ENV > defaultConfig
func getConfigPath() string {

	//TODO: FIX THE CONFIG PATH PARCING
	defaultPath := "./config/local.yaml"
	var configPath string

	flag.StringVar(&configPath, "config", "", "Path to your config file (.yaml)")
	flag.Parse()

	if configPath != "" {
		return configPath
	} else if os.Getenv("CONFIG_PATH") != "" {
		return os.Getenv("CONFIG_PATH")
	} else {
		return defaultPath
	}

}
