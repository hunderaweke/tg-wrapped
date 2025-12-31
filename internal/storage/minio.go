package storage

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	apperrors "github.com/hunderaweke/tg-unwrapped/internal/errors"
	"github.com/hunderaweke/tg-unwrapped/internal/logger"
)

const (
	defaultTimeout = 30 * time.Second
)

type MinioClient struct {
	BucketName string
	Client     *minio.Client
}

func NewMinioBucket(bucket string) (*MinioClient, error) {
	log := logger.With("operation", "NewMinioBucket", "bucket", bucket)

	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" {
		return nil, apperrors.NewConfigError("MINIO_ENDPOINT", apperrors.ErrInvalidConfig)
	}

	accessKey := os.Getenv("MINIO_ACCESS_ID")
	if accessKey == "" {
		return nil, apperrors.NewConfigError("MINIO_ACCESS_ID", apperrors.ErrInvalidConfig)
	}

	secretKey := os.Getenv("MINIO_SECRET_ID")
	if secretKey == "" {
		return nil, apperrors.NewConfigError("MINIO_SECRET_ID", apperrors.ErrInvalidConfig)
	}

	token := os.Getenv("MINIO_TOKEN")

	clnt, err := minio.New(endpoint, &minio.Options{
		Creds: credentials.NewStaticV4(accessKey, secretKey, token),
	})
	if err != nil {
		log.Error("Failed to create Minio client", "error", err)
		return nil, fmt.Errorf("%w: %v", apperrors.ErrMinioConnection, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	exists, err := clnt.BucketExists(ctx, bucket)
	if err != nil {
		log.Error("Failed to check bucket existence", "error", err)
		return nil, fmt.Errorf("%w: %v", apperrors.ErrMinioConnection, err)
	}

	if !exists {
		logger.Info("Bucket does not exist, creating it", "bucket", bucket)
		if err := clnt.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			log.Error("Failed to create bucket", "error", err)
			return nil, fmt.Errorf("%w: %v", apperrors.ErrMinioConnection, err)
		}
		logger.Info("Created new bucket", "bucket", bucket)
	}

	logger.Info("Minio client initialized successfully",
		"endpoint", endpoint,
		"bucket", bucket)

	return &MinioClient{
		Client:     clnt,
		BucketName: bucket,
	}, nil
}

func (m *MinioClient) GeneratePresignedURL(objectName string, expiryTime time.Duration) (string, error) {
	log := logger.With("operation", "GeneratePresignedURL", "object", objectName)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	presignedURL, err := m.Client.PresignedPutObject(ctx, m.BucketName, objectName, expiryTime)
	if err != nil {
		log.Error("Failed to generate presigned URL", "error", err)
		return "", fmt.Errorf("failed to generate presignedUrl: %v", err)
	}

	log.Debug("Generated presigned URL", "expiry", expiryTime)
	return presignedURL.String(), nil
}

func (m *MinioClient) UploadProfile(fileName string, profile bytes.Buffer, contentType string) error {
	log := logger.With("operation", "UploadProfile", "filename", fileName)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	reader := bytes.NewReader(profile.Bytes())
	_, err := m.Client.PutObject(ctx, m.BucketName, fileName, reader, int64(profile.Len()), minio.PutObjectOptions{
		Expires:     time.Now().Add(48 * time.Hour),
		ContentType: contentType,
	})
	if err != nil {
		log.Error("Failed to upload object", "error", err)
		return fmt.Errorf("%w: %v", apperrors.ErrUploadFailed, err)
	}

	log.Debug("Object uploaded successfully",
		"bucket", m.BucketName,
		"size", profile.Len(),
		"content_type", contentType)

	return nil
}

func (m *MinioClient) GenerateAccessURL(objectName string, expiryTime time.Duration) (string, error) {
	log := logger.With("operation", "GenerateAccessURL", "object", objectName)

	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	presignedURL, err := m.Client.PresignedGetObject(ctx, m.BucketName, objectName, expiryTime, url.Values{})
	if err != nil {
		log.Error("Failed to generate access URL", "error", err)
		return "", fmt.Errorf("error creating get link for the object: %v %v", err, objectName)
	}

	log.Debug("Generated access URL", "expiry", expiryTime)
	return presignedURL.String(), nil
}
