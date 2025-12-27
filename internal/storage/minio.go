package storage

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	BucketName string
	Client     *minio.Client
}

func NewMinioBucket(bucket string) (*MinioClient, error) {
	clnt, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("error getting minio client: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	exists, errBucketExists := clnt.BucketExists(ctx, bucket)
	if errBucketExists == nil && !exists {
		log.Printf("bucket do not exist creating it ... %v", bucket)
		if err := clnt.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
	} else if errBucketExists != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", errBucketExists)
	}

	return &MinioClient{
		Client:     clnt,
		BucketName: bucket,
	}, nil
}

func (m *MinioClient) GeneratePresignedURL(objectName string, expiryTime time.Duration) (string, error) {
	ctx := context.Background()
	presignedURL, err := m.Client.PresignedPutObject(ctx, m.BucketName, objectName, expiryTime)
	if err != nil {
		return "", fmt.Errorf("failed to generate presignedUrl: %v", err)
	}
	return presignedURL.String(), nil
}

func (m *MinioClient) UploadProfile(fileName string, profile bytes.Buffer, contentType string) error {
	reader := bytes.NewReader(profile.Bytes())
	_, err := m.Client.PutObject(context.Background(), m.BucketName, fileName, reader, int64(profile.Len()), minio.PutObjectOptions{
		Expires:     time.Now().Add(48 * time.Hour),
		ContentType: contentType,
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *MinioClient) GenerateAccessURL(objectName string, expiryTime time.Duration) (string, error) {
	ctx := context.Background()
	presignedURL, err := m.Client.PresignedGetObject(ctx, m.BucketName, objectName, expiryTime, url.Values{})
	if err != nil {
		return "", fmt.Errorf("error creating get link for the object: %v %v", err, objectName)
	}
	return presignedURL.String(), nil
}

func getClient() (*minio.Client, error) {
	return minio.New(os.Getenv("MINIO_ENDPOINT"), &minio.Options{
		Creds: credentials.NewStaticV4(os.Getenv("MINIO_ACCESS_ID"), os.Getenv("MINIO_SECRET_ID"), os.Getenv("MINIO_TOKEN")),
	})
}
