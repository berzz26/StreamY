package upload

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterRoutes(app *fiber.App) {

	app.Post("/upload", UploadVideo)
}

func UploadVideo(c *fiber.Ctx) error {

	file, err := c.FormFile("video")
	if err != nil {
		return fiber.NewError(
			fiber.StatusBadRequest,
			"video file required",
		)
	}

	videoID := uuid.New().String()

	err = os.MkdirAll("./uploads", os.ModePerm)
	if err != nil {
		return err
	}

	extension := filepath.Ext(file.Filename)

	localPath := fmt.Sprintf(
		"./uploads/%s%s",
		videoID,
		extension,
	)

	err = c.SaveFile(file, localPath)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"video_id": videoID,
		"path":     localPath,
		"status":   "uploaded",
	})
}