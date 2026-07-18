package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Rest struct {
	Host    string        `yaml:"host"`
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type Grpc struct {
	Port    int           `yaml:"port"`
	Timeout time.Duration `yaml:"timeout"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	SslMode  string `yaml:"ssl_mode"`
	Password string `env:"DB_PASS" env-required:"true"`
}

type Redis struct {
	Addr string `yaml:"addr"`
}

type AiClient struct {
	Url string `yaml:"base_url"`
	Key string `env:"API_KEY" env-required:"true"`
}

type Pipeline struct {
	Count      int `yaml:"worker_count"`
	BufferSize int `yaml:"buffer_size"`
}

type Config struct {
	Rest     Rest     `yaml:"rest_service"`
	Grpc     Grpc     `yaml:"grpc_service"`
	Db       DB       `yaml:"db"`
	Redis    Redis    `yaml:"redis"`
	AiClient AiClient `yaml:"ai_client"`
	Pipeline Pipeline `yaml:"pipeline"`
}

func MustLoad() *Config {
	_ = godotenv.Load(".env")

	cfg := Config{}
	err := cleanenv.ReadConfig("config/config.yml", &cfg)
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}
	return &cfg
}
