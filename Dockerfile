# Dockerfile
# FROM golang:1.24.2-alpine3.20
FROM golang:1.25.4-alpine AS builder

# Install build dependencies
RUN apk update && apk add --no-cache gcc libc-dev libwebp-dev

WORKDIR /app

COPY src/ .

# set env for chai
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=arm64

# download packages
RUN go mod download

# Build the Go application for Linux (using Alpine)
RUN GOOS=linux GOARCH=arm64 go build -o main .
# start alpine for ffmpeg
FROM alpine:latest
RUN apk add --no-cache ffmpeg libwebp tzdata ca-certificates
COPY --from=0 /app/main /usr/bin/main

# Expose the port your application will run on
EXPOSE 4400

# Run the application using gin for live-reloading
CMD ["main"]