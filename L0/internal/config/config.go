package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	DsnPQ      DsnPQ      `env-prefix:"DSN_"`
	HTTPServer HTTPServer `env-prefix:"HTTP_"`
	Cache      Cache      `env-prefix:"CACHE_"`
	Kafka      Kafka      `env-prefix:"KAFKA_"`
}

type DsnPQ struct {
	Port     int    `env:"PORT" env-default:"5432"`
	User     string `env:"USER" env-default:"admin"`
	Password string `env:"PASSWORD" env-default:"adm_123"`
	Name     string `env:"NAME" env-default:"myapp"`
	Host     string `env:"HOST" env-default:"localhost"`
}

type Kafka struct {
	Brokers  string `env:"BROKERS" env-default:"localhost:9092"`
	Topic    string `env:"TOPIC" env-default:"user-events"`
	DLQTopic string `env:"DLQ_TOPIC" env-default:"user-events-dlq"`
	GroupID  string `env:"GROUP_ID" env-default:"user-api-group"`
}

type Cache struct {
	TTL      time.Duration `env:"TTL" env-default:"60s"`
	Capacity int           `env:"CAPACITY" env-default:"10"`
}

type HTTPServer struct {
	Address     string        `env:"ADDR" env-default:"localhost:8081"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"4s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad(path string) *Config {
	const op = "MustLoad"
	if err := LoadEnv(path); err != nil {
		fmt.Printf("%+v\n", err)
		log.Println("no .env file found, reading from environment variables", "error", err, "op", op)
	}

	cfg, err := Load()
	if err != nil {
		log.Fatalf("%s: failed to load config: %v", op, err)
	}
	return cfg

}

func Load() (*Config, error) {
	const op = "Load"
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &cfg, nil
}

func LoadEnv(path string) error {
	const op = "LoadEnv"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Println("no .env file found, reading from environment variables")
		return nil
	}
	if err := godotenv.Load(path); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
