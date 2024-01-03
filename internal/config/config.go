package config

import "C"
import (
	"flag"
	"log/slog"

	"github.com/caarlos0/env/v6"
)

type HTTPServer struct {
	Host        string     `json:"host" env:"RUN_ADDRESS" envDefault:":8080"`
	DSN         string     `json:"dsn,omitempty" env:"DATABASE_URI" envDefault:"postgres://postgres:Xer_0101@localhost/gophermart?sslmode=disable"`
	Accrual     string     `json:"accrual" env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://localhost:8081"`
	LogLevel    slog.Level `json:"log_level" env:"LOG_LEVEL"`
	SecretToken string     `json:"secret_token" env:"SECRET_TOKEN" envDefault:"sky-go-mart"`
}

var Cfg HTTPServer

func MakeConfig() error {
	flag.StringVar(&Cfg.Host, "a", Cfg.Host, "Host HTTP-server")
	flag.StringVar(&Cfg.DSN, "d", Cfg.DSN, "Connect to database")
	flag.StringVar(&Cfg.Accrual, "r", Cfg.Accrual, "Accrual system address")
	flag.Var(&LogLevelValue{&Cfg.LogLevel}, "l", "Log level (debug, info, warn, error)")
	flag.StringVar(&Cfg.SecretToken, "k", Cfg.SecretToken, "Secret key for the token")
	flag.Parse()
	return env.Parse(&Cfg)
}
