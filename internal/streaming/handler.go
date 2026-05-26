package streaming

import (
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/", HealthCheck)
	app.Get("/stream/:video/*", StreamVideo)
}

func HealthCheck(c *fiber.Ctx) error {
	return c.SendString("Server is up")
}

func StreamVideo(c *fiber.Ctx) error {

	videoId := c.Params("video")
	file := c.Params("*")

	fullPath := fmt.Sprintf(
		"./storage/processed/%s/%s",
		videoId,
		file,
	)

	ext := filepath.Ext(fullPath)

	switch ext {

	case ".m3u8":
		c.Set("Content-Type", "application/vnd.apple.mpegurl")

	case ".ts":
		c.Set("Content-Type", "video/mp2t")
	}

	// Sends file as stream internally
	return c.SendFile(fullPath)
}