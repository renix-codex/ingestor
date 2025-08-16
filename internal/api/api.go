package api

import (
	"time"

	"github.com/renix-codex/ingestor/internal/ingest"
)

// API struct holds all the dependent services required for APIs
// API is the application-facing facade. All callers (HTTP, CLI, gRPC) go through this.
type API struct {
	ing *ingest.Service
}

func New(ing *ingest.Service) *API {
	return &API{ing: ing}
}

// Health responds with the health status of the app, as a map[string]interface{}
func (api *API) Health() interface{} {
	payload := map[string]interface{}{
		"app":       "ingestor",
		"startedAt": time.Now().Format(time.RFC3339),
		"status":    "ok",
	}
	return payload
}
