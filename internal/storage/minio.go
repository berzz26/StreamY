package storage

import (
	"context"
	"log"

	"github.com/berzz26/StreamY/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
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
		return nil, err
	}

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
