package streaming

import (
	"fmt"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App) {
	app.Get("/stream/:video/*", StreamVideo)
}
func StreamVideo(c *fiber.Ctx) error {
	videoId := c.Params("video")

	file := c.Params("*")

	fullPath := fmt.Sprintf("./storage/processed/%s/%s", videoId, file)

	ext := filepath.Ext(fullPath)
	switch ext {
	//set req headers
	case ".m3u8":
		c.Set("Content-Type", "application/vnd.apple.mpegurl")

	case ".ts":
		c.Set("Content-Type", "video/mp2t")
	}
	//FOR TEST ONLY (fiber opens the file internally and starts streaming the bytes)
	return c.SendFile(fullPath)
}
