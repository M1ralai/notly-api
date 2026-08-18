package storage

import (
	"context"
	"fmt"
	"mime/multipart"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOAdapter implements StorageProvider backed by a MinIO (S3-compatible) server.
type MinIOAdapter struct {
	client     *minio.Client
	bucketName string
	endpoint   string
}

// NewMinIOAdapter constructs a MinIOAdapter from environment variables.
//
// Expected env vars:
//
//	MINIO_ENDPOINT   – e.g. "localhost:9000"
//	MINIO_ACCESS_KEY
//	MINIO_SECRET_KEY
//	MINIO_BUCKET     – e.g. "notly-media"
//	MINIO_USE_SSL    – "true" | "false"  (default: false)
func NewMinIOAdapter() (*MinIOAdapter, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("minio: MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY and MINIO_BUCKET must be configured")
	}
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: failed to create client: %w", err)
	}

	return &MinIOAdapter{
		client:     client,
		bucketName: bucket,
		endpoint:   endpoint,
	}, nil
}

// EnsureBucket creates the configured bucket if it does not already exist.
func (a *MinIOAdapter) EnsureBucket() error {
	ctx := context.Background()
	exists, err := a.client.BucketExists(ctx, a.bucketName)
	if err != nil {
		return fmt.Errorf("minio: bucket check failed: %w", err)
	}
	if !exists {
		if err := a.client.MakeBucket(ctx, a.bucketName, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("minio: bucket creation failed: %w", err)
		}
	}
	return nil
}

// Upload stores the given multipart.File in MinIO under the provided filename.
// It returns the full object URL that can be stored in the database.
func (a *MinIOAdapter) Upload(file multipart.File, filename, contentType string) (string, error) {
	ctx := context.Background()

	// Detect size so we can stream efficiently; -1 accepts unknown size.
	size := int64(-1)
	if seeker, ok := file.(interface {
		Seek(int64, int) (int64, error)
	}); ok {
		end, err := seeker.Seek(0, 2) // seek to end
		if err == nil {
			size = end
			_, _ = seeker.Seek(0, 0) // rewind
		}
	}

	_, err := a.client.PutObject(ctx, a.bucketName, filename, file, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio: upload failed for %q: %w", filename, err)
	}

	// Build the public URL.  In production this would be served through Nginx.
	url := fmt.Sprintf("http://%s/%s/%s", a.endpoint, a.bucketName, filename)
	return url, nil
}

// Delete removes an object from the bucket by its object key.
func (a *MinIOAdapter) Delete(objectKey string) error {
	ctx := context.Background()
	err := a.client.RemoveObject(ctx, a.bucketName, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio: delete failed for %q: %w", objectKey, err)
	}
	return nil
}

func (a *MinIOAdapter) Download(ctx context.Context, objectKey string) (*StoredObject, error) {
	info, err := a.client.StatObject(ctx, a.bucketName, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio: stat failed for %q: %w", objectKey, err)
	}

	object, err := a.client.GetObject(ctx, a.bucketName, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio: download failed for %q: %w", objectKey, err)
	}

	return &StoredObject{
		Body:        object,
		ContentType: info.ContentType,
		Size:        info.Size,
	}, nil
}
