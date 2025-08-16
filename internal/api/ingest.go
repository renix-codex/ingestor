package api

import (
	"context"

	"github.com/renix-codex/ingestor/internal/models"
)

// IngestOnce triggers a single ingestion run.
func (a *API) IngestOnce(ctx context.Context) (int, error) {
	return a.ing.IngestOnce(ctx)
}

// QueryByUser returns enriched posts for a user.
func (a *API) QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error) {
	return a.ing.QueryByUser(ctx, userID)
}
