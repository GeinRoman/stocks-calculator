package server

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-playground/validator"
	"github.com/spf13/viper"
)

type serverConfig struct {
	Port                 int    `mapstructure:"PORT"                    validate:"required"`
	DBUsername           string `mapstructure:"DATABASE_USER"           validate:"required"`
	DBPassword           string `mapstructure:"DATABASE_PASSWORD"       validate:"required"`
	MaxConnections       int    `mapstructure:"max_connections"         validate:"required"`
	JwtKey               string `mapstructure:"jwt_key"                 validate:"required"`
	JwtExpirationTimeMin int    `mapstructure:"jwt_expiration_time_min" validate:"required"`
}

func loadConfig() (serverConfig, error) {
	var config serverConfig

	// load .env for develop
	loadDotEnv()

	viper.SetConfigFile("config.yaml")
	viper.AutomaticEnv()

	for _, key := range []string{"PORT", "DATABASE_USER", "DATABASE_PASSWORD"} {
		viper.BindEnv(key)
	}

	if err := viper.ReadInConfig(); err != nil {
		return config, fmt.Errorf("Failed to read config file: %w", err)
	}

	if err := viper.Unmarshal(&config); err != nil {
		return config, fmt.Errorf("Failed to unmarshal config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&config); err != nil {
		return config, fmt.Errorf("Invalid config: %w", err)
	}

	return config, nil
}

func loadDotEnv() {
	filename := ".env"
	data, err := os.ReadFile(filename)
	if err != nil {
		return
	}

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		os.Setenv(key, value)
	}
}
