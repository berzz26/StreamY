package main

import (
	"github.com/berzz26/StreamY/internal/config"
	"github.com/berzz26/StreamY/internal/storage"
	"log"
)

func main() {
	cfg := config.LoadConfig()
	client, err := storage.NewMinioClient(*cfg)

	if err != nil {
		panic(err)
	}

	log.Println("MinIO connected:", client.EndpointURL())

	err = storage.UploadFile(
		client,

		cfg.MinioBucket,

		"originals/test.mp4",

		"./storage/originals/test.mp4",

		"video/mp4",
	)

	if err != nil {
		panic(err)
	}
}
