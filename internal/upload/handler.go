package upload

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/config"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	repo *repository.VideoRepository
}

func NewHandler(
	repo *repository.VideoRepository,
) *Handler {

	return &Handler{
		repo: repo,
	}
}

func (h *Handler) RegisterRoutes(
	app *fiber.App,
) {

	app.Post("/upload", h.UploadVideo)
}

func (h *Handler) UploadVideo(
	c *fiber.Ctx,
) error {
	cfg := config.LoadConfig()
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

	video := models.Video{
		ID:           videoID,
		Title:        file.Filename,
		Status:       models.StatusUploaded,
		OriginalPath: localPath,
		OriginalSize: file.Size,
	}

	err = h.repo.CreateVideo(video)
	if err != nil {
		return err
	}
	videoUrl := fmt.Sprintf("%s/stream/%s/index.m3u8",cfg.HostURL, videoID)
	return c.JSON(fiber.Map{
		"video_id": videoID,
		"status":   models.StatusUploaded,
		"url" : videoUrl,
	})
}
