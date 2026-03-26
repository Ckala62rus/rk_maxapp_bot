package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string `yaml:"env" env:"ENV" env-default:"development" env-required:"true"`
	ConfigFile string `yaml:"config_file" env:"config_file" env-default:"config" env-required:"true"`
	//EnableTracing  bool   `yaml:"enable_tracing" env:"ENABLE_TRACING" env-default:"false"`
	//OtelEndpoint   string `yaml:"otel_endpoint" env:"OTEL_ENDPOINT" env-default:"jaeger:4318"`
	HttpServer  HttpServer  `yaml:"http_server"`
	MinioConfig MinioConfig `yaml:"minio"`
	//DatabaseConfig `yaml:"database"`
	//RedisConfig    `yaml:"redis"`
	//SecurityConfig `yaml:"security"`
}

type HttpServer struct {
	Address     string        `yaml:"address" evn-default:"localhost"`
	Timeout     time.Duration `yaml:"timeout" env-default:"4s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	Port        string        `yaml:"port" env-default:"3000"`
}

type DatabaseConfig struct {
	Host     string `yaml:"postgres_host"`
	Port     string `yaml:"postgres_port"`
	User     string `yaml:"postgres_user"`
	Password string `yaml:"postgres_password"`
	Db       string `yaml:"postgres_db"`
}

type MinioConfig struct {
	Host     string `yaml:"minio_host"`
	User     string `yaml:"minio_user"`
	Password string `yaml:"minio_password"`
	SSL      bool   `yaml:"minio_ssl"`
}

//type RedisConfig struct {
//	Host string `yaml:"redis_host"`
//	Port string `yaml:"redis_port"`
//	Db   int    `yaml:"redis_db"`
//}

//type SecurityConfig struct {
//	JWTSigningKey string `yaml:"jwt_signing_key" env:"JWT_SIGNING_KEY" env-required:"true"`
//	JwtTokenTTL   string `yaml:"jwt_token_ttl" env:"JWT_TOKEN_TTL" env-default:"10"`
//}

func MustLoad(configPath string) *Config {
	//configPath := os.Getenv("CONFIG_PATH")

	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}

	var cfg Config

	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	// Валидация критически важных секретов
	//if cfg.SecurityConfig.JWTSigningKey == "" {
	//	log.Fatal("JWT_SIGNING_KEY is required but not set")
	//}

	//if len(cfg.SecurityConfig.JWTSigningKey) < 32 {
	//	log.Fatal("JWT_SIGNING_KEY must be at least 32 characters long")
	//}

	return &cfg
}
