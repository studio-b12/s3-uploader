package uploader

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Uploader struct {
	client         *s3.Client
	bucket         string
	runningUploads *sync.Map
}

type S3Options struct {
	BaseEndpoint    *string
	AccessKeyID     *string
	SecretAccessKey *string
	SessionToken    string
}

func NewS3Uploader(region string, bucket string, options S3Options) (*S3Uploader, error) {
	var client *s3.Client

	if options.BaseEndpoint != nil || options.AccessKeyID != nil || options.SecretAccessKey != nil {
		opts := s3.Options{
			AppID:        "s3-uploader/0.1.0",
			Region:       region,
			BaseEndpoint: options.BaseEndpoint,
		}

		if options.AccessKeyID != nil && options.SecretAccessKey != nil {
			opts.Credentials = credentials.StaticCredentialsProvider{Value: aws.Credentials{
				AccessKeyID:     *options.AccessKeyID,
				SecretAccessKey: *options.SecretAccessKey,
				SessionToken:    options.SessionToken,
			}}
		}

		client = s3.New(opts)
	} else {
		cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(region))
		if err != nil {
			return nil, err
		}

		client = s3.NewFromConfig(cfg)
	}

	t := &S3Uploader{
		client:         client,
		bucket:         bucket,
		runningUploads: &sync.Map{},
	}
	return t, nil
}

func (t *S3Uploader) UploadFile(filePath string, key string) error {
	if cancelFunc, ok := t.runningUploads.Load(key); ok {
		cancelFunc.(context.CancelFunc)()
		slog.Warn("upload canceled", "path", filePath, "key", key)
	}

	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	ctx, cancelFunc := context.WithCancel(context.TODO())
	t.runningUploads.Store(key, cancelFunc)
	defer t.runningUploads.Delete(key)

	_, err = t.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &t.bucket,
		Key:    aws.String(key),
		Body:   f,
	})

	return err
}
