package main

import (
	"log"

	"github.com/berzz26/StreamY/internal/transcoder"
)

func main() {
	log.Println("worker started")
	//pass the inputPath and the outputdir
	err := transcoder.ProcessVideo(
		"./storage/originals/test.mp4",
		"./storage/processed",
	)

	if err != nil {
		panic(err)
	}

	log.Println("transcoding completed")
}