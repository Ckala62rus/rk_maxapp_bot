// Package config holds app configuration structures and loader.
package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config describes full application configuration.
type Config struct {
	Env        string    `yaml:"env" env:"ENV" env-default:"development"`
	Log        LogConfig `yaml:"log"`
	HttpServer HttpServer `yaml:"http_server"`
	Postgres   PostgresConfig `yaml:"postgres"`
	MSSQL      MSSQLConfig    `yaml:"mssql"`
	Max        MaxConfig      `yaml:"max"`
}

// LogConfig controls logging settings.
type LogConfig struct {
	Level string `yaml:"level" env:"LOG_LEVEL" env-default:"info"`
}

// HttpServer describes HTTP server settings.
type HttpServer struct {
	Address     string        `yaml:"address" env-default:"0.0.0.0"`
	Port        string        `yaml:"port" env-default:"3000"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
}

// PostgresConfig describes PostgreSQL connection settings.
type PostgresConfig struct {
	Host     string `yaml:"host" env:"POSTGRES_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"POSTGRES_PORT" env-default:"5432"`
	User     string `yaml:"user" env:"POSTGRES_USER" env-default:"postgres"`
	Password string `yaml:"password" env:"POSTGRES_PASSWORD" env-default:"postgres"`
	Database string `yaml:"database" env:"POSTGRES_DB" env-default:"app"`
	SSLMode  string `yaml:"sslmode" env:"POSTGRES_SSLMODE" env-default:"disable"`
}

// MSSQLConfig describes MSSQL (DAX) connection settings.
type MSSQLConfig struct {
	Host     string `yaml:"host" env:"DAX_MSSQL_HOST" env-default:"localhost"`
	Port     string `yaml:"port" env:"DAX_MSSQL_PORT" env-default:"1433"`
	User     string `yaml:"user" env:"DAX_MSSQL_USER" env-default:"sa"`
	Password string `yaml:"password" env:"DAX_MSSQL_PASSWORD" env-default:""`
	Database string `yaml:"database" env:"DAX_MSSQL_DB" env-default:""`
	Encrypt  string `yaml:"encrypt" env:"DAX_MSSQL_ENCRYPT" env-default:"disable"`
}

// MaxConfig holds MAX bot configuration.
type MaxConfig struct {
	BotToken string `yaml:"bot_token" env:"BOT_TOKEN" env-required:"true"`
	// AllowMockInitData enables dev-only endpoint to generate initData.
	AllowMockInitData bool `yaml:"allow_mock_initdata" env:"ENABLE_INITDATA_MOCK" env-default:"false"`
}

// MustLoad loads config from file or panics on error.
func MustLoad(path string) *Config {
	if path == "" {
		// CONFIG_PATH is required to avoid loading random files in container.
		log.Fatal("CONFIG_PATH is not set")
	}

	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
		// Stop early if config cannot be parsed.
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
