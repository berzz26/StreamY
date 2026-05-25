package config

import (
	"log"
	"os"
)

type Config struct {
	DatabaseUrl string
}

func LoadDb() *Config {
	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		log.Fatal("DATABASE_URL is missing")
	}
	return &Config{
		DatabaseUrl: dbUrl,
	}
}
