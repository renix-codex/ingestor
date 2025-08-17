package ingest

import (
	"context"

	"github.com/renix-codex/ingestor/internal/models"
)

type CollectorPort interface {
	Fetch(ctx context.Context) ([]models.Post, error)
}

type StorePort interface {
	Upsert(ctx context.Context, items []models.EnrichedPost) error
	QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error)
	QueryRecent(ctx context.Context, limit, offset int) ([]models.EnrichedPost, error)
}
