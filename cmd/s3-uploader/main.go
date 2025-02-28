package main

import (
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/studio-b12/s3-uploader/pkg/uploader"
	"github.com/studio-b12/s3-uploader/pkg/watcher"

	"github.com/alexflint/go-arg"
	"github.com/gammazero/workerpool"
	"github.com/joho/godotenv"
)

type Args struct {
	Directory       string     `arg:"--directory,env:S3U_DIRECTORY,required" help:"Directory to watch for files to upload"`
	IntervalSeconds int        `arg:"--interval,env:S3U_INTERVAL" default:"10" help:"Check interval for file changes in seconds"`
	LogLevel        slog.Level `arg:"--log-level,env:S3U_LOGLEVEL" default:"info" help:"Log level"`

	ParallelUploads int `arg:"--parallel-uploads,env:S3U_PARALLELUPLOADS" default:"5" help:"Maximum number of parallel uploads"`
	UploadQueueSize int `arg:"--upload-queue-size,env:S3U_UPLOADQUEUESIZE" default:"50" help:"Size for upload queue; should larger as the expected amount of files that change per check cycle"`

	S3Region          string  `arg:"--s3-region,env:S3U_S3_REGION,required" help:"S3 region of the upload bucket"`
	S3Bucket          string  `arg:"--s3-bucket,env:S3U_S3_BUCKET,required" help:"S3 bucket to upload to"`
	S3Endpoint        *string `arg:"--s3-endpoint,env:S3U_S3_ENDPOINT" help:"S3 endpoint URL"`
	S3AccessKeyId     *string `arg:"--s3-accesskeyid,env:S3U_S3_ACCESSKEYID" help:"S3 access key ID"`
	S3SecretAccessKey *string `arg:"--s3-secretacceskey,env:S3U_S3_SECRETACCESSKEY" help:"S3 secret access key"`
	S3SessionToken    string  `arg:"--s3-sessiontoken,env:S3U_S3_SESSIONTOKEN" help:"S3 session token"`
}

func main() {
	godotenv.Load()

	var args Args
	arg.MustParse(&args)

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: args.LogLevel}))
	slog.SetDefault(logger)

	u := uploader.NewS3Uploader(args.S3Region, args.S3Bucket, uploader.S3Options{
		BaseEndpoint:    args.S3Endpoint,
		AccessKeyID:     args.S3AccessKeyId,
		SecretAccessKey: args.S3SecretAccessKey,
		SessionToken:    args.S3SessionToken,
	})

	w := watcher.NewPollingWatcher(os.DirFS(args.Directory), time.Duration(args.IntervalSeconds)*time.Second, args.UploadQueueSize)

	slog.Info("start watching ...", "path", args.Directory, "interval-seconds", args.IntervalSeconds)
	w.Start()
	defer w.Stop()

	wp := workerpool.New(args.ParallelUploads)

	for file := range w.Results() {
		fullPath := filepath.Join(args.Directory, file.Path)
		wp.Submit(func() {
			key := path.Clean(file.Path)

			slog.Info("uploading file ...", "path", fullPath, "key", key)

			err := u.UploadFile(fullPath, key)
			if err != nil {
				slog.Error("upload failed", "path", fullPath, "key", key, "err", err)
				return
			}

			slog.Info("upload finished", "path", fullPath, "key", key)
		})
	}
}
