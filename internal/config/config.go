package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	// HTTP server
	ListenAddr  string        // e.g. ":8080"
	HTTPTimeout time.Duration // e.g. 10s (for upstream API)

	// Upstream source
	SourceURL  string // e.g. https://jsonplaceholder.typicode.com/posts
	SourceName string // e.g. "placeholder_api"

	// Postgres (explicit pieces)
	PGHost     string // e.g. "localhost" or "postgres" when running in compose
	PGPort     int    // e.g. 5432
	PGUser     string // e.g. "app"
	PGPassword string // e.g. "app"
	PGDatabase string // e.g. "ingestor"
	PGSSLMode  string // e.g. "disable" locally, "require" in cloud
}

// BuildDSN composes a keyword/value DSN compatible with pgxpool.
func (c Config) BuildDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.PGHost, c.PGPort, c.PGUser, c.PGPassword, c.PGDatabase, c.PGSSLMode,
	)
}

func FromEnv() Config {
	c := Config{}

	c.ListenAddr = getenv("HTTP_LISTEN_ADDR", ":8080")

	if d, err := time.ParseDuration(getenv("HTTP_TIMEOUT", "10s")); err == nil {
		c.HTTPTimeout = d
	} else {
		c.HTTPTimeout = 10 * time.Second
	}

	c.SourceURL = getenv("SOURCE_URL", "https://jsonplaceholder.typicode.com/posts")
	c.SourceName = getenv("SOURCE_NAME", "placeholder_api")

	// Postgres pieces
	c.PGHost = getenv("PG_HOST", "postgres")
	c.PGPort = getenvi("PG_PORT", 5432)
	c.PGUser = getenv("PG_USER", "app")
	c.PGPassword = getenv("PG_PASSWORD", "app")
	c.PGDatabase = getenv("PG_DATABASE", "ingestor")
	c.PGSSLMode = getenv("PG_SSLMODE", "disable")

	return c
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func getenvi(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		var iv int
		_, err := fmt.Sscanf(v, "%d", &iv)
		if err == nil {
			return iv
		}
	}
	return def
}
