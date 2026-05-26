package storage

import (
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