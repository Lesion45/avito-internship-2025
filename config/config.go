package config

import (
	"fmt"
	"time"

	"github.com/caarlos0/env/v8"
)

type Config struct {
	Env      string        `env:"ENV,required"`
	TokenTTL time.Duration `env:"TOKEN_TTL,required"`
	Salt     string        `env:"SALT,required"`
	PgDSN    string        `env:"POSTGRES_DSN,required"`
	RedisDSN string        `env:"REDIS_DSN,required"`
}

type Kafka struct {
	Host  string `env:"BROKER_HOST,required"`
	Topic string `env:"BROKER_TOPIC,required"`
}

// MustLoad loads configuration from config.yaml
// Throw a panic if the config doesn't exist or if there is an error reading the config.
func MustLoad() *Config {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		panic(fmt.Sprintf("Failed to load config: %s", err))
	}

	return &cfg
}
