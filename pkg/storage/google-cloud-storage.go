package storage

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
)

type GoogleCloudStorageClientWrapper struct {
	client *storage.Client
}

func (csc *GoogleCloudStorageClientWrapper) Client(ctx context.Context) error {
	c, err := storage.NewClient(ctx)

	if err != nil {
		return err
	}

	csc.client = c
	return nil
}

func (csc *GoogleCloudStorageClientWrapper) GetObject(ctx context.Context, bucketName, fn string) ([]byte, error) {
	log.Printf("getting object %v from bucket %v", fn, bucketName)
	bkt := csc.client.Bucket(bucketName)

	rc, err := bkt.Object(fn).NewReader(ctx)

	if err != nil {
		log.Printf("failed getting object: %v, %v", fn, err)
		return nil, err
	}

	defer rc.Close()

	file, err := io.ReadAll(rc)
	if err != nil {
		log.Printf("failed reading object: %v, %v", fn, err)
		return nil, err
	}

	return file, nil
}
