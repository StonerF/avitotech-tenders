package config

import (
	"log"
	"os"
	"time"
)

type Config struct {
	StoragePath string
	HttpServer
}

type HttpServer struct {
	Address      string
	Timeout      time.Duration
	Idle_timeout time.Duration
}

func MustLoad() *Config {
	Server_adress := os.Getenv("SERVER_ADDRESS")
	if Server_adress == "" {
		log.Fatal("Server address is empty")
	}
	var Cfg Config
	Cfg.Address = Server_adress
	str_path := os.Getenv("POSTGERS_CONN")
	Cfg.StoragePath = str_path
	Cfg.Timeout = 4 * time.Second
	Cfg.Idle_timeout = 60 * time.Second

	return &Cfg

}
