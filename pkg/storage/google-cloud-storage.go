package storage

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
	"github.com/wizact/te-reo-bot/pkg/entities"
	"github.com/wizact/te-reo-bot/pkg/logger"
)

type GoogleCloudStorageClientWrapper struct {
	client *storage.Client
	logger logger.Logger
}

// NewGoogleCloudStorageClientWrapper creates a new wrapper with logger
func NewGoogleCloudStorageClientWrapper(logger logger.Logger) *GoogleCloudStorageClientWrapper {
	return &GoogleCloudStorageClientWrapper{
		logger: logger,
	}
}

func (csc *GoogleCloudStorageClientWrapper) Client(ctx context.Context) error {
	c, err := storage.NewClient(ctx)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to create Google Cloud Storage client")
		appErr = appErr.WithContext("operation", "client_initialization")

		csc.logger.ErrorWithStack(appErr, "Google Cloud Storage client initialization failed",
			logger.String("operation", "client_initialization"),
			logger.String("error_type", "client_creation"),
		)
		return appErr
	}

	csc.client = c
	csc.logger.Info("Google Cloud Storage client initialized successfully",
		logger.String("operation", "client_initialization"),
	)
	return nil
}

func (csc *GoogleCloudStorageClientWrapper) GetObject(ctx context.Context, bucketName, fn string) ([]byte, error) {
	csc.logger.Info("Getting object from Google Cloud Storage",
		logger.String("operation", "get_object"),
		logger.String("bucket_name", bucketName),
		logger.String("object_name", fn),
	)

	bkt := csc.client.Bucket(bucketName)
	rc, err := bkt.Object(fn).NewReader(ctx)

	if err != nil {
		appErr := entities.NewAppError(err, 404, "Failed to get object from Google Cloud Storage")
		appErr = appErr.WithContext("operation", "get_object")
		appErr = appErr.WithContext("bucket_name", bucketName)
		appErr = appErr.WithContext("object_name", fn)

		csc.logger.ErrorWithStack(appErr, "Failed to create object reader",
			logger.String("operation", "get_object"),
			logger.String("bucket_name", bucketName),
			logger.String("object_name", fn),
			logger.String("error_type", "object_reader_creation"),
		)
		return nil, appErr
	}

	defer rc.Close()

	file, err := io.ReadAll(rc)
	if err != nil {
		appErr := entities.NewAppError(err, 500, "Failed to read object data from Google Cloud Storage")
		appErr = appErr.WithContext("operation", "read_object_data")
		appErr = appErr.WithContext("bucket_name", bucketName)
		appErr = appErr.WithContext("object_name", fn)

		csc.logger.ErrorWithStack(appErr, "Failed to read object data",
			logger.String("operation", "read_object_data"),
			logger.String("bucket_name", bucketName),
			logger.String("object_name", fn),
			logger.String("error_type", "object_data_reading"),
		)
		return nil, appErr
	}

	csc.logger.Info("Successfully retrieved object from Google Cloud Storage",
		logger.String("operation", "get_object"),
		logger.String("bucket_name", bucketName),
		logger.String("object_name", fn),
		logger.Int("data_size_bytes", len(file)),
	)

	return file, nil
}
