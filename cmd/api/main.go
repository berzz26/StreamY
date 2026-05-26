package main

import (
	// "context"
	// "fmt"

	// "time"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/berzz26/StreamY/internal/upload"
	// "github.com/berzz26/StreamY/internal/database"
	"log"

	"github.com/berzz26/StreamY/internal/streaming"
	"github.com/gofiber/fiber/v2"
)

func main() {

	cfg := config.LoadConfig()
	// db := database.New(cfg.DatabaseUrl)

	// defer db.Close()

	// ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	// defer cancel()

	// var version string

	// err := db.DB.QueryRow(ctx, "SELECT version()").Scan(&version)
	// if err != nil {
	// 	panic(err)

	// }

	// fmt.Println(version)

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024, // 1GB
	})
	streaming.RegisterRoutes(app)
	upload.RegisterRoutes(app)

	log.Printf("Api server up on %s", cfg.Port)

	err := app.Listen(":" + cfg.Port)
	if err != nil {
		panic(err)
	}

}
