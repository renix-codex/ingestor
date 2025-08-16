# syntax=docker/dockerfile:1.7

FROM golang:1.23-alpine AS builder
WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
COPY vendor/ ./vendor/
COPY . .

ARG TARGET=./main.go
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -trimpath -ldflags="-s -w" -o /out/ingestor ${TARGET}

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /out/ingestor /app/ingestor
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/ingestor"]
