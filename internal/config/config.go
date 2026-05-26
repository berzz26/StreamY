package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
)

type Config struct {
	DatabaseUrl string
	Port        string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}
	dbUrl := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	if dbUrl == "" {
		log.Fatal("DATABASE_URL is missing")
	}

	if port == "" {
		log.Fatal("Port is missing")
	}
	return &Config{
		DatabaseUrl: dbUrl,
		Port:        port,
	}
}
