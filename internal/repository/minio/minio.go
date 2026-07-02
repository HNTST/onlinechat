package minio

import (
	"context"
	config "discord/internal/core/config"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	cfg config.Config_Minio
	mc  *minio.Client
}

func NewMinio(cfg config.Config_Minio) *MinioClient {
	return &MinioClient{
		cfg: cfg,
	}
}

func (m *MinioClient) InitMinio() error {
	ctx := context.Background()

	client, err := minio.New(m.cfg.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(m.cfg.RootUser, m.cfg.RootPassword, ""),
		Secure: true,
	})
	if err != nil {
		return err
	}
	m.mc = client

	// Проверка наличия бакета и его создание, если не существует
	exists, err := m.mc.BucketExists(ctx, m.cfg.BucketName)
	if err != nil {
		return err
	}
	if !exists {
		err := m.mc.MakeBucket(ctx, m.cfg.BucketName, minio.MakeBucketOptions{Region: ""})
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *MinioClient) UploadFile(ctx context.Context, file_name, content_type string, reader io.Reader, size int64) (string, error) {
	opts := minio.PutObjectOptions{
		ContentType: content_type,
	}
	info, err := m.mc.PutObject(ctx, m.cfg.BucketName, file_name, reader, size, opts)
	if err != nil {
		return "", err
	}
	return info.Key, nil
}
