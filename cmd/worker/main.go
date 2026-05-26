package main

import (
	"context"
	"log"
	"time"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/berzz26/StreamY/internal/database"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/transcoder"
	"github.com/minio/minio-go/v7"
)

func main() {

	cfg := config.LoadConfig()
	db := database.New(cfg.DatabaseUrl)

	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var version string

	err := db.DB.QueryRow(ctx, "SELECT version()").Scan(&version)
	if err != nil {
		panic(err)

	}

	log.Println(version)
	if err != nil {
		panic(err)
	}

	videoRepo := repository.NewVideoRepository(db.DB)

	worker := transcoder.NewWorker(
		videoRepo,
		&minio.Client{},
		cfg.MinioBucket,
	)

	worker.Start()
}
