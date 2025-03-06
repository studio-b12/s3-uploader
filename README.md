# s3-uploader

Simple tool to watch for files in a directory and upload them to an S3 bucket.

The given directory will be watched recursively. Every file that is new or has not been updated between two polling
cycles will be uploaded to the given S3 bucket.

## Limitations

This tool is primarily designed to handle files that increase in size over time, like logs or benchmark outputs. The
file watcher just looks for changes in file modification time and file size, therefore, if the file system does not
support modification time and if the files only change in content and not in size, the change will not be detected.

## Docker Image

```
ghcr.io/studio-b12/s3-uploader
```

## Configuration

All configuration variables can be viewed via the `--help` flag. Configuration can either be passed via command line
arguments or via environment variables.

```
Usage: s3-uploader --directory DIRECTORY [--interval INTERVAL] [--log-level LOG-LEVEL] [--parallel-uploads PARALLEL-UPLOADS] [--upload-queue-size UPLOAD-QUEUE-SIZE] --s3-region S3-REGION --s3-bucket S3-BUCKET [--s3-endpoint S3-ENDPOINT] [--s3-accesskeyid S3-ACCESSKEYID] [--s3-secretacceskey S3-SECRETACCESKEY] [--s3-sessiontoken S3-SESSIONTOKEN]

Options:
  --directory DIRECTORY
                         Directory to watch for files to upload [env: S3U_DIRECTORY]
  --interval INTERVAL    Check interval for file changes in seconds [default: 10, env: S3U_INTERVAL]
  --log-level LOG-LEVEL
                         Log level [default: info, env: S3U_LOGLEVEL]
  --parallel-uploads PARALLEL-UPLOADS
                         Maximum number of parallel uploads [default: 5, env: S3U_PARALLELUPLOADS]
  --upload-queue-size UPLOAD-QUEUE-SIZE
                         Size for upload queue; should larger as the expected amount of files that change per check cycle [default: 50, env: S3U_UPLOADQUEUESIZE]
  --delete-after-upload
                         Delete files on local disk after successful upload [env: S3U_DELETEAFTERUPLOAD]
  --s3-region S3-REGION
                         S3 region of the upload bucket [env: S3U_S3_REGION]
  --s3-bucket S3-BUCKET
                         S3 bucket to upload to [env: S3U_S3_BUCKET]
  --s3-endpoint S3-ENDPOINT
                         S3 endpoint URL [env: S3U_S3_ENDPOINT]
  --s3-accesskeyid S3-ACCESSKEYID
                         S3 access key ID [env: S3U_S3_ACCESSKEYID]
  --s3-secretacceskey S3-SECRETACCESKEY
                         S3 secret access key [env: S3U_S3_SECRETACCESSKEY]
  --s3-sessiontoken S3-SESSIONTOKEN
                         S3 session token [env: S3U_S3_SESSIONTOKEN]
  --help, -h             display this help and exit
```