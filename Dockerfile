FROM golang:1-alpine AS build
WORKDIR /build
COPY cmd/ cmd/
COPY pkg/ pkg/
COPY go.mod go.mod
COPY go.sum go.sum
RUN go build -o s3-uploader ./cmd/s3-uploader/main.go

FROM alpine:latest AS final
COPY --from=build /build/s3-uploader /usr/bin/s3-uploader
ENV S3U_DIRECTORY="/upload"
ENTRYPOINT [ "/usr/bin/s3-uploader" ]
