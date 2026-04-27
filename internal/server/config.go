package server

import (
	"github.com/spf13/viper"
)

type serverConfig struct {
	Port                 int    `mapstructure:"port"`
	MaxConnections       int    `mapstructure:"max_connections"`
	JwtKey               string `mapstructure:"jwt_key"`
	JwtExpirationTimeMin int    `mapstructure:"jwt_expiration_time_min"`
	DbParams             struct {
		Address  string `mapstructure:"address"`
		Username string `mapstructure:"username"`
		Password string `mapstructure:"password"`
		SslMode  string `mapstructure:"ssl_mode"`
	} `mapstructure:"db_params"`
}

func loadConfig() (serverConfig, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	var config serverConfig
	err := viper.ReadInConfig()
	if err != nil {
		return config, err
	}

	err = viper.Unmarshal(&config)
	return config, nil
}
