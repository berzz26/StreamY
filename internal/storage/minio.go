package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

func NewMinioClient(cfg config.Config) (*minio.Client, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds: credentials.NewStaticV4(
			cfg.MinioAccessKey,
			cfg.MinioSecretKey,
			"",
		),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	log.Printf("MinIO client initialized: endpoint=%s secure=%t",
		cfg.MinioEndpoint,
		cfg.MinioUseSSL,
	)

	return client, nil
}

func UploadFile(
	client *minio.Client,
	bucket string,
	objectName string,
	filePath string,
	contentType string,
) error {
	info, err := client.FPutObject(
		context.Background(),
		bucket,
		objectName,
		filePath,
		minio.PutObjectOptions{
			ContentType: contentType,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"failed uploading %s to %s/%s: %w",
			filePath,
			bucket,
			objectName,
			err,
		)
	}

	log.Printf(
		"uploaded %s size=%d",
		info.Key,
		info.Size,
	)

	return nil
}

func UploadDirectory(
	client *minio.Client,
	bucket string,
	localDir string,
	objectPrefix string,
) error {
	type uploadJob struct {
		path        string
		objectName  string
		contentType string
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	jobs := make(chan uploadJob)

	var g errgroup.Group

	// Fixed number of upload workers.
	for i := 0; i < 4; i++ {
		workerID := i

		g.Go(func() error {
			for {
				select {
				case <-ctx.Done():
					return ctx.Err()

				case job, ok := <-jobs:
					if !ok {
						return nil
					}

					_, err := client.FPutObject(
						ctx,
						bucket,
						job.objectName,
						job.path,
						minio.PutObjectOptions{
							ContentType: job.contentType,
						},
					)

					if err != nil {
						return fmt.Errorf(
							"worker %d: failed uploading %s to %s/%s: %w",
							workerID,
							job.path,
							bucket,
							job.objectName,
							err,
						)
					}

					log.Printf(
						"worker %d: uploaded %s",
						workerID,
						job.objectName,
					)
				}
			}
		})
	}

	// Discover files and feed them to workers.
	walkErr := filepath.Walk(
		localDir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if info.IsDir() {
				return nil
			}

			// Stop walking if an upload worker already failed.
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			relativePath, err := filepath.Rel(localDir, path)
			if err != nil {
				return fmt.Errorf(
					"failed getting relative path for %s: %w",
					path,
					err,
				)
			}

			objectName := filepath.Join(
				objectPrefix,
				relativePath,
			)

			// MinIO object names should always use /.
			objectName = filepath.ToSlash(objectName)

			contentType := "application/octet-stream"

			switch filepath.Ext(path) {
			case ".m3u8":
				contentType = "application/vnd.apple.mpegurl"

			case ".ts":
				contentType = "video/mp2t"
			}

			job := uploadJob{
				path:        path,
				objectName:  objectName,
				contentType: contentType,
			}

			select {
			case <-ctx.Done():
				return ctx.Err()

			case jobs <- job:
				return nil
			}
		},
	)

	close(jobs)

	// If walking failed, cancel workers.
	if walkErr != nil {
		cancel()

		// Wait for workers to exit before returning.
		_ = g.Wait()

		return fmt.Errorf(
			"failed walking upload directory %s: %w",
			localDir,
			walkErr,
		)
	}

	// Wait for all workers.
	if err := g.Wait(); err != nil {
		return fmt.Errorf(
			"directory upload failed for %s: %w",
			localDir,
			err,
		)
	}

	log.Printf(
		"directory upload completed: %s -> %s/%s",
		localDir,
		bucket,
		objectPrefix,
	)

	return nil
}