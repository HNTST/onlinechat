package files_repository

import (
	"context"
	config "discord/internal/core/config"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioStorage struct {
	mc  *minio.Client
	cfg config.Config_Minio
}

func NewMinioStorage(cfg config.Config_Minio) (*MinioStorage, error) {
	client, err := minio.New(cfg.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.RootUser, cfg.RootPassword, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}
	return &MinioStorage{mc: client, cfg: cfg}, nil
}

func (m *MinioStorage) Upload(ctx context.Context, fileName, contentType string, reader io.Reader, size int64) (string, error) {
	opts := minio.PutObjectOptions{ContentType: contentType}
	info, err := m.mc.PutObject(ctx, m.cfg.BucketName, fileName, reader, size, opts)
	if err != nil {
		return "", err
	}
	return info.Key, nil
}
