package config

import (
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	GrpcPort string
	Postgres PostgresConfig

	UserServiceGrpcPort string
	UserServiceHost     string
}

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Database string
}

func Load(path string) Config {
	godotenv.Load(path + "/.env")

	conf := viper.New()
	conf.AutomaticEnv()

	cfg := Config{
		GrpcPort: conf.GetString("GRPC_PORT"),
		Postgres: PostgresConfig{
			Host:     conf.GetString("POSTGRES_HOST"),
			Port:     conf.GetString("POSTGRES_PORT"),
			User:     conf.GetString("POSTGRES_USER"),
			Password: conf.GetString("POSTGRES_PASSWORD"),
			Database: conf.GetString("POSTGRES_DATABASE"),
		},
		UserServiceHost:     conf.GetString("USER_SERVICE_HOST"),
		UserServiceGrpcPort: conf.GetString("USER_SERVICE_GRPC_PORT"),
	}

	return cfg
}
