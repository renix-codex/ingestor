# --- builder (Alpine) ---
FROM golang:1.22-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# Static binary so it runs anywhere (no CGO deps)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/ingestor ./main.go

# --- runtime (Alpine) ---
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=build /out/ingestor /app/ingestor
EXPOSE 8080
# Non-root for safety
RUN adduser -D -u 10001 appuser
USER appuser
ENTRYPOINT ["/app/ingestor"]
