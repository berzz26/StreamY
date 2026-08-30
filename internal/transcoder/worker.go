package transcoder

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/berzz26/StreamY/internal/models"
	"github.com/berzz26/StreamY/internal/repository"
	"github.com/berzz26/StreamY/internal/storage"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
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

	if info, err := os.Stat(video.OriginalPath); err == nil {
		video.OriginalSize = info.Size()
	}

	run := models.Run{
		ID:             uuid.New().String(),
		VideoID:        video.ID,
		OriginalSize:   video.OriginalSize,
		StartedAt:      time.Now(),
		RenditionTimes: map[string]int64{},
	}

	fail := func(errMessage string) {
		run.Status = models.StatusFailed
		run.ErrorMessage = errMessage
	}

	defer func() {
		run.FinishedAt = time.Now()
		run.TotalMS = run.FinishedAt.Sub(run.StartedAt).Milliseconds()

		if run.Status == "" {
			run.Status = models.StatusFailed
		}

		if err := w.repo.CreateRun(&run); err != nil {
			logger.Error().
				Err(err).
				Str("videoID", video.ID).
				Msg("failed to save pipeline run")
		}
	}()

	defer func() {
		if err := os.Remove(video.OriginalPath); err != nil {
			logger.Warn().Err(err).Str("videoID", video.ID).Msg("cleanup original failed")
		}
		if err := os.RemoveAll(outputDir); err != nil {
			logger.Warn().Err(err).Str("videoID", video.ID).Msg("cleanup processed failed")
		}
		logger.Info().Str("videoID", video.ID).Msg("cleaned temp files")
	}()

	probeStart := time.Now()

	probe, err := ProbeVideo(video.OriginalPath)

	if err != nil {
		run.ProbeMS = time.Since(probeStart).Milliseconds()

		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("video probing failed")

		fail(err.Error())
		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	run.ProbeMS = time.Since(probeStart).Milliseconds()

	probe.VideoID = video.ID

	if err := w.repo.CreateVideoProbe(probe); err != nil {
		logger.Error().Err(err).Str("videoID", video.ID).Msg("failed to save video probe")
	}
	if probe.FormatDuration != nil {
		if err := w.repo.UpdateVideoDuration(video.ID, *probe.FormatDuration); err != nil {
			logger.Error().Err(err).Str("videoID", video.ID).Msg("failed to save video duration")
		}
		video.DurationSeconds = *probe.FormatDuration
		run.DurationSeconds = *probe.FormatDuration
	}

	renditions := PlanRenditions(probe)

	run.RenditionsCount = len(renditions)

	logger.Info().
		Str("videoID", video.ID).
		Int("renditions", len(renditions)).
		Msg("planned renditions")

	start := time.Now()

	var g errgroup.Group

	for i := range renditions {
		rendition := renditions[i]
		resDir := fmt.Sprintf("%dp", rendition.Height)
		resolutionDir := filepath.Join(outputDir, resDir)

		encodeStart := time.Now()

		if err := EncodeRendition(
			&rendition,
			outputDir,
			video.OriginalPath,
		); err != nil {
			run.RenditionTimes[fmt.Sprintf("%d", rendition.Height)] = time.Since(encodeStart).Milliseconds()

			logger.Error().
				Err(err).
				Str("videoID", video.ID).
				Msg("transcoding failed")

			_ = g.Wait()
			fail(err.Error())
			w.repo.MarkVideoFailed(
				video.ID,
				err.Error(),
			)
			return
		}

		run.RenditionTimes[fmt.Sprintf("%d", rendition.Height)] = time.Since(encodeStart).Milliseconds()

		g.Go(func() error {
			return storage.UploadDirectory(
				w.minio,
				w.bucket,
				resolutionDir,
				"processed/"+video.ID+"/"+resDir,
			)
		})
	}

	transcodeDur := time.Since(start)

	run.TranscodeMS = transcodeDur.Milliseconds()

	logger.Info().
		Str("videoID", video.ID).
		Dur("transcode_time", transcodeDur).
		Msg("transcoding completed")

	uploadStart := time.Now()

	if err := g.Wait(); err != nil {
		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("minio upload failed")

		fail(err.Error())
		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	uploadDur := time.Since(uploadStart)

	run.UploadMS = uploadDur.Milliseconds()

	logger.Info().
		Str("videoID", video.ID).
		Dur("upload_time", uploadDur).
		Msg("uploaded assets to minio")

	if err := CreateMasterPlaylist(renditions, outputDir); err != nil {
		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("failed to create master playlist")

		fail(err.Error())
		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	if err := storage.UploadFile(
		w.minio,
		w.bucket,
		"processed/"+video.ID+"/index.m3u8",
		filepath.Join(outputDir, "index.m3u8"),
		"application/vnd.apple.mpegurl",
	); err != nil {
		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Msg("failed to upload master playlist")

		fail(err.Error())
		w.repo.MarkVideoFailed(
			video.ID,
			err.Error(),
		)
		return
	}

	processedPath := "processed/" + video.ID

	dbStart := time.Now()

	err = w.repo.UpdateVideoStatus(
		video.ID,
		models.StatusProcessed,
		processedPath,
	)

	dbDur := time.Since(dbStart)

	run.DbMS = dbDur.Milliseconds()

	if err != nil {

		logger.Error().
			Err(err).
			Str("videoID", video.ID).
			Dur("db_time", dbDur).
			Msg("failed to update status")

		fail(err.Error())

		return
	}

	run.Status = models.StatusProcessed

	totalDur := time.Since(start)

	logger.Info().
		Str("OriginalSize", fmt.Sprintf("%v ", video.OriginalSize)).
		Str("DurationSeconds", fmt.Sprintf("%v ", video.DurationSeconds)).
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
