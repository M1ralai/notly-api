package storage

import (
	"context"
	"io"
	"mime/multipart"
)

type StoredObject struct {
	Body        io.ReadCloser
	ContentType string
	Size        int64
}

// StorageProvider is the abstraction for any object-storage backend.
// It knows nothing about notes or business logic – it only moves bytes.
type StorageProvider interface {
	// Upload stores the given file under the specified object key and returns
	// the public/presigned URL that can later be persisted in the database.
	Upload(file multipart.File, filename, contentType string) (url string, err error)

	// Delete removes an object by its key (filename / path inside the bucket).
	Delete(objectKey string) error

	// Download opens an object stream by its key.
	Download(ctx context.Context, objectKey string) (*StoredObject, error)

	// EnsureBucket verifies that the configured bucket exists; it creates it
	// if it is missing.  Should be called once at application startup.
	EnsureBucket() error
}
