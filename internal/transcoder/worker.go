package transcoder

import (
	"log"
	"os"
	"time"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/storage"

	"github.com/minio/minio-go/v7"
)

type Worker struct {
	repo   *repository.VideoRepository
	minio  *minio.Client
	bucket string
}

func NewWorker(
	repo *repository.VideoRepository,
	minioClient *minio.Client,
	bucket string,
) *Worker {

	return &Worker{
		repo:   repo,
		minio:  minioClient,
		bucket: bucket,
	}
}

func (w *Worker) Start() {

	log.Println("worker started")

	for {

		video, err := w.repo.ClaimNextVideo()
		if err != nil {

			log.Println(err)

			time.Sleep(5 * time.Second)

			continue
		}

		if video == nil {

			time.Sleep(5 * time.Second)

			continue
		}

		log.Printf(
			"claimed video %s",
			video.ID,
		)

		w.processVideo(video)
	}
}

func (w *Worker) processVideo(video *models.Video) {

	outputDir := "./processed/" + video.ID

	defer func() {
		if err := os.Remove(video.OriginalPath); err != nil {
			log.Printf("cleanup original: %v", err)
		}
		if err := os.RemoveAll(outputDir); err != nil {
			log.Printf("cleanup processed: %v", err)
		}
		log.Printf("cleaned temp files for %s", video.ID)
	}()

	err := ProcessVideo(
		video.OriginalPath,
		outputDir,
	)

	if err != nil {

		log.Println(err)

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)

		return
	}

	log.Println("transcoding completed")

	err = storage.UploadDirectory(
		w.minio,

		w.bucket,

		outputDir,

		"processed/"+video.ID,
	)

	if err != nil {

		log.Println(err)

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)

		return
	}

	log.Println("uploaded assets to minio")

	err = w.repo.UpdateVideoStatus(
		video.ID,
		models.StatusProcessed,
	)

	if err != nil {

		log.Println(err)

		return
	}

	log.Printf(
		"video %s processed successfully",
		video.ID,
	)
}
