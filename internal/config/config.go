package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseUrl string
	Port        string

	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file not found")
	}

	dbUrl := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")
	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	minioBucket := os.Getenv("MINIO_BUCKET")
	minioUseSSLStr := os.Getenv("MINIO_USE_SSL")

	if dbUrl == "" {
		log.Fatal("DATABASE_URL is missing")
	}

	if port == "" {
		log.Fatal("PORT is missing")
	}

	if minioEndpoint == "" {
		log.Fatal("MINIO_ENDPOINT is missing")
	}

	if minioAccessKey == "" {
		log.Fatal("MINIO_ACCESS_KEY is missing")
	}

	if minioSecretKey == "" {
		log.Fatal("MINIO_SECRET_KEY is missing")
	}

	if minioBucket == "" {
		log.Fatal("MINIO_BUCKET is missing")
	}

	useSSL := false
	if minioUseSSLStr != "" {
		var parseErr error
		useSSL, parseErr = strconv.ParseBool(minioUseSSLStr)
		if parseErr != nil {
			log.Fatalf("MINIO_USE_SSL must be a boolean value: %v", parseErr)
		}
	}

	return &Config{
		DatabaseUrl:    dbUrl,
		Port:           port,
		MinioEndpoint:  minioEndpoint,
		MinioAccessKey: minioAccessKey,
		MinioSecretKey: minioSecretKey,
		MinioBucket:    minioBucket,
		MinioUseSSL:    useSSL,
	}
}
