package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddr       string
	PostgresUser     string
	PostgresPasswrod string
	PostgresHost     string
	PostgresPort     string
	PostgresDB       string
	RedisHost        string
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	AppConfig = &Config{
		ServerAddr:       viper.GetString("SERVER_ADDR"),
		PostgresUser:     viper.GetString("DB_APP_USER"),
		PostgresPasswrod: viper.GetString("DB_APP_PASSWORD"),
		PostgresHost:     viper.GetString("POSTGRES_HOST"),
		PostgresPort:     viper.GetString("POSTGRES_PORT"),
		PostgresDB:       viper.GetString("POSTGRES_DB"),
		RedisHost:        viper.GetString("REDIS_HOST"),
	}
}
