package transcoder

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/storage"

	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
)

var logger = zerolog.New(
	zerolog.ConsoleWriter{Out: os.Stderr},
).With().Timestamp().Logger()

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

	logger.Info().Msg("worker started")

	for {

		video, err := w.repo.ClaimNextVideo()
		if err != nil {

			logger.Warn().Err(err).Msg("failed to claim video")

			time.Sleep(5 * time.Second)

			continue
		}

		if video == nil {

			time.Sleep(5 * time.Second)

			continue
		}

		logger.Info().
			Str("videoID", video.ID).
			Msg("claimed video")

		w.processVideo(video)
	}
}

func (w *Worker) processVideo(video *models.Video) {

	outputDir := "./processed/" + video.ID

	defer func() {
		if err := os.Remove(video.OriginalPath); err != nil {
			logger.Warn().Err(err).Str("videoID", video.ID).Msg("cleanup original failed")
		}
		if err := os.RemoveAll(outputDir); err != nil {
			logger.Warn().Err(err).Str("videoID", video.ID).Msg("cleanup processed failed")
		}
		logger.Info().Str("videoID", video.ID).Msg("cleaned temp files")
	}()

	probe, err := ProbeVideo(video.OriginalPath)

	if err != nil {
		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("video probing failed")

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	probe.VideoID = video.ID

	renditions := PlanRenditions(probe)

	logger.Info().
		Str("videoID", video.ID).
		Int("renditions", len(renditions)).
		Msg("planned renditions")

	start := time.Now()

	// err = ProcessVideo(
	// 	video.OriginalPath,
	// 	outputDir,
	// )

	for i := range renditions {
		err = EncodeRendition(
			&renditions[i],
			outputDir,
			video.OriginalPath,
		)

		if err != nil {
			logger.Error().
				Err(err).
				Str("videoID", video.ID).
				Msg("transcoding failed")

			w.repo.MarkVideoFailed(
				video.ID,
				err.Error(),
			)
			return
		}
	}

	transcodeDur := time.Since(start)
	if err := CreateMasterPlaylist(renditions, outputDir); err != nil {
		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("failed to create master playlist")

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	if err != nil {

		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Dur("transcode_time", transcodeDur).
			Msg("transcoding failed")

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)

		return
	}

	logger.Info().
		Str("videoID", video.ID).
		Dur("transcode_time", transcodeDur).
		Msg("transcoding completed")

	uploadStart := time.Now()
	processedPath := "processed/" + video.ID

	err = storage.UploadDirectory(
		w.minio,

		w.bucket,

		outputDir,

		processedPath,
	)

	uploadDur := time.Since(uploadStart)

	if err != nil {

		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Dur("upload_time", uploadDur).
			Msg("minio upload failed")

		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)

		return
	}

	logger.Info().
		Str("videoID", video.ID).
		Dur("upload_time", uploadDur).
		Msg("uploaded assets to minio")

	dbStart := time.Now()

	err = w.repo.UpdateVideoStatus(
		video.ID,
		models.StatusProcessed,
		processedPath,
	)

	dbDur := time.Since(dbStart)

	if err != nil {

		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Dur("db_time", dbDur).
			Msg("failed to update status")

		return
	}

	totalDur := time.Since(start)

	logger.Info().
		Str("videoID", video.ID).
		Dur("transcode_time", transcodeDur).
		Dur("upload_time", uploadDur).
		Dur("db_time", dbDur).
		Dur("total_time", totalDur).
		Msg("video processed successfully")
}

func CreateMasterPlaylist(
	renditions []models.VideoRendition,
	outputDir string,
) error {

	masterPath := filepath.Join(outputDir, "index.m3u8")

	file, err := os.Create(masterPath)
	if err != nil {
		return err
	}
	defer file.Close()

	fmt.Fprintln(file, "#EXTM3U")
	fmt.Fprintln(file, "#EXT-X-VERSION:3")

	for _, rendition := range renditions {
		fmt.Fprintf(
			file,
			"#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n",
			rendition.Bitrate,
			rendition.Width,
			rendition.Height,
		)

		fmt.Fprintf(
			file,
			"%dp/index.m3u8\n",
			rendition.Height,
		)
	}

	return nil
}
