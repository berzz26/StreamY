package streaming

import (
	"context"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/minio/minio-go/v7"
)

type Handler struct {
	minio  *minio.Client
	bucket string
}

func NewHandler(
	minioClient *minio.Client,
	bucket string,
) *Handler {

	return &Handler{
		minio:  minioClient,
		bucket: bucket,
	}
}

func (h *Handler) RegisterRoutes(
	app *fiber.App,
) {

	app.Get("/", h.HealthCheck)

	app.Get(
		"/stream/:video/*",
		h.StreamVideo,
	)
}

func (h *Handler) HealthCheck(
	c *fiber.Ctx,
) error {

	return c.SendString("server is up")
}

func (h *Handler) StreamVideo(
	c *fiber.Ctx,
) error {

	videoID := c.Params("video")

	file := c.Params("*")

	objectName := filepath.Join(
		"processed",
		videoID,
		file,
	)

	object, err := h.minio.GetObject(
		context.Background(),

		h.bucket,

		objectName,

		minio.GetObjectOptions{},
	)

	if err != nil {
		return err
	}

	ext := filepath.Ext(file)

	switch ext {

	case ".m3u8":
		c.Set(
			"Content-Type",
			"application/vnd.apple.mpegurl",
		)

	case ".ts":
		c.Set(
			"Content-Type",
			"video/mp2t",
		)
	}

	return c.SendStream(object)
}