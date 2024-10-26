package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	ServerAddr         string
	AccessTokenSecret  string
	RefreshTokenSecret string
	PostgresUser       string
	PostgresPasswrod   string
	PostgresHost       string
	PostgresPort       string
	PostgresDB         string
	RedisHost          string
	SmtpHost           string
	SmtpPort           string
	SmtpUsername       string
	SmtpPassword       string
	AppURL             string
}

var AppConfig *Config

func LoadConfig() {
	viper.SetConfigFile(".env")
	viper.ReadInConfig()
	viper.AutomaticEnv()
	AppConfig = &Config{
		ServerAddr:         viper.GetString("SERVER_ADDR"),
		AccessTokenSecret:  viper.GetString("ACCESS_TOKEN_SECRET"),
		RefreshTokenSecret: viper.GetString("REFRESH_TOKEN_SECRET"),
		PostgresUser:       viper.GetString("DB_APP_USER"),
		PostgresPasswrod:   viper.GetString("DB_APP_PASSWORD"),
		PostgresHost:       viper.GetString("POSTGRES_HOST"),
		PostgresPort:       viper.GetString("POSTGRES_PORT"),
		PostgresDB:         viper.GetString("POSTGRES_DB"),
		RedisHost:          viper.GetString("REDIS_HOST"),
		SmtpHost:           viper.GetString("SMTP_HOST"),
		SmtpPort:           viper.GetString("SMTP_PORT"),
		SmtpUsername:       viper.GetString("SMTP_USERNAME"),
		SmtpPassword:       viper.GetString("SMTP_PASSWORD"),
		AppURL:             viper.GetString("APP_URL"),
	}
}
