package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Port     string `yaml:"port" env:"PORT" env-default:"5432"`
	Host     string `yaml:"host" env:"HOST" env-default:"localhost"`
	Name     string `yaml:"name" env:"NAME" env-default:"postgres"`
	User     string `yaml:"user" env:"USER" env-default:"user"`
	Password string `yaml:"password" env:"PASSWORD"`
	HttpServer
}

type HttpServer struct {
	Address      string        `yaml:"address" env:"ADDERSS" env-default:"localhost:8080"`
	Timeout      time.Duration `yaml:"timeout" env:"TIMEOUT" env-default:"4s"`
	Idle_timeout time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" env-default:"60s"`
}

func MustLoad() *Config {
	var Cfg Config
	if err := cleanenv.ReadConfig("../config/config.yaml", &Cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &Cfg

}
