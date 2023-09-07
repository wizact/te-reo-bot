package wotd

import (
	"context"
	"io"

	"cloud.google.com/go/storage"
)

type CloudStorageClientWrapper struct {
	client *storage.Client
}

func (csc *CloudStorageClientWrapper) Client(ctx context.Context) error {
	c, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	csc.client = c
	return nil
}

func (csc *CloudStorageClientWrapper) GetObject(ctx context.Context, bucketName, fn string) ([]byte, error) {
	bkt := csc.client.Bucket(bucketName)

	rc, err := bkt.Object(fn).NewReader(ctx)

	if err != nil {
		return nil, err
	}

	defer rc.Close()

	file, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return file, nil
}
