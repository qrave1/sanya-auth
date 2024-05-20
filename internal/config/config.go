package config

import (
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	Server   server
	Postgres postgres
}

type server struct {
	Port   string `env:"SERVER_PORT" envDefault:"8080"`
	Secret string `env:"SERVER_SECRET" envDefault:"mirea_the_best"`
}

type postgres struct {
	Host     string `env:"POSTGRES_HOST" envDefault:"postgres"`
	Port     string `env:"POSTGRES_PORT" envDefault:"5432"`
	Username string `env:"POSTGRES_USERNAME" envDefault:"user"`
	Password string `env:"POSTGRES_PASSWORD" envDefault:"password"`
	Database string `env:"POSTGRES_DATABASE" envDefault:"auth"`
}

func NewConfig() *Config {
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Errorf("error parsing config: %w", err))
	}

	return &cfg
}
