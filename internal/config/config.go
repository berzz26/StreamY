package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DatabaseUrl string
}

func LoadDb() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}
	dbUrl := os.Getenv("DATABASE_URL")

	if dbUrl == "" {
		log.Fatal("DATABASE_URL is missing")
	}
	return &Config{
		DatabaseUrl: dbUrl,
	}
}
