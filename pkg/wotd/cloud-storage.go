package wotd

import (
	"context"
	"io/ioutil"

	"cloud.google.com/go/storage"
	"github.com/kelseyhightower/envconfig"
	ent "github.com/wizact/te-reo-bot/pkg/entities"
)

func NewCloudStorageClient() (*storage.Client, *ent.AppError) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		return nil, &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	return client, nil
}

func GetObject(gsc *storage.Client, fn string) ([]byte, *ent.AppError) {
	mbc, e := getMediaBucketName()
	if e != nil {
		return nil, e
	}

	bkt := gsc.Bucket(mbc)

	rc, err := bkt.Object(fn).NewReader(context.Background())

	if err != nil {
		return nil, &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	defer rc.Close()

	file, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	return file, nil
}

func getMediaBucketName() (string, *ent.AppError) {
	var s StorageConfig
	err := envconfig.Process("tereobot", &s)
	if err != nil {
		return "nil", &ent.AppError{Error: err, Code: 500, Message: "Failed to acquire image"}
	}

	return s.BucketName, nil
}

type StorageConfig struct {
	BucketName string
}
