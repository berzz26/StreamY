package storage

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"golang.org/x/sync/errgroup"
)

func NewMinioClient(cfg config.Config) (*minio.Client, error) {
	transport, err := minio.DefaultTransport(cfg.MinioUseSSL)
	if err != nil {
		return nil, err
	}
	transport.MaxIdleConnsPerHost = 64
	transport.MaxIdleConns = 128

	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:     credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure:    cfg.MinioUseSSL,
		Transport: transport,
	})
	if err != nil {
		return nil, err
	}
	log.Printf("Connected to Minio")
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
		return err
	}

	log.Printf(
		"uploaded %s size %d",
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

	jobs := make(chan uploadJob)

	var g errgroup.Group

	// Start a fixed number of upload workers.
	for i := 0; i < 7; i++ {
		g.Go(func() error {
			for job := range jobs {

				_, err := client.FPutObject(
					context.Background(),
					bucket,
					job.objectName,
					job.path,
					minio.PutObjectOptions{
						ContentType: job.contentType,
					},
				)

				if err != nil {
					return err
				}

				log.Printf("uploaded %s", job.objectName)
			}

			return nil
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

			relativePath, err := filepath.Rel(localDir, path)
			if err != nil {
				return err
			}

			objectName := filepath.Join(
				objectPrefix,
				relativePath,
			)

			contentType := "application/octet-stream"

			switch filepath.Ext(path) {
			case ".m3u8":
				contentType = "application/vnd.apple.mpegurl"
			case ".ts":
				contentType = "video/mp2t"
			}

			jobs <- uploadJob{
				path:        path,
				objectName:  objectName,
				contentType: contentType,
			}

			return nil
		},
	)

	close(jobs)

	if walkErr != nil {
		return walkErr
	}

	return g.Wait()
}
