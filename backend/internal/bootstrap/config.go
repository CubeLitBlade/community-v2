package bootstrap

import (
	"errors"
	"os"
)

type Config struct {
	Addr        string
	DatabaseURL string
	SnowflakeID int64
}

func LoadConfig() (Config, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return Config{}, errors.New("DATABASE_URL is unset")
	}

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		Addr:        addr,
		DatabaseURL: dsn,
		SnowflakeID: 1,
	}, nil
}
