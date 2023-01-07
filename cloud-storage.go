package main

import (
	"context"

	"cloud.google.com/go/storage"
)

func NewCloudStorageClient() (*storage.Client, *AppError) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, &AppError{Error: err, Code: 500, Message: "Failed acquire photo"}
	}

	return client, nil
}
