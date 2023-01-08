package main

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
)

func newCloudStorageClient() (*storage.Client, *AppError) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, &AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	return client, nil
}

func getObject(gsc *storage.Client, fn string) ([]byte, *AppError) {
	bkt := gsc.Bucket("te-reo-bot-images")

	rc, err := bkt.Object(fn).NewReader(context.Background())

	if err != nil {
		return nil, &AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	defer rc.Close()

	file, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, &AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	return file, nil
}
