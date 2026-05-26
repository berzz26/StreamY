package main

import (
	"context"
	"fmt"

	"time"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/berzz26/StreamY/internal/database"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/upload"
	"log"

	"github.com/berzz26/StreamY/internal/streaming"
	"github.com/gofiber/fiber/v2"
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

	fmt.Println(version)

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024,

		ErrorHandler: func(c *fiber.Ctx, err error) error {

			// log actual internal error
			log.Printf("request failed: %v", err)

			// generic response to client
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "internal server error",
			})
		},
	})
	streaming.RegisterRoutes(app)

	videoRepo := repository.NewVideoRepository(db.DB)

	uploadHandler := upload.NewHandler(videoRepo)
	uploadHandler.RegisterRoutes(app)

	log.Printf("Api server up on %s", cfg.Port)

	err2 := app.Listen(":" + cfg.Port)
	if err2 != nil {
		panic(err2)
	}

}
