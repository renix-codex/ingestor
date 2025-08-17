package ingest

import (
	"context"
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

type Service struct {
	store     StorePort
	collector CollectorPort
	source    string
	now       func() time.Time
}

// IngestOnce fetches posts from the collector, enriches them, and stores them in the database.
func (s *Service) IngestOnce(ctx context.Context) (int, error) {
	posts, err := s.collector.Fetch(ctx)
	if err != nil {
		return 0, err
	}
	enriched := Enrich(posts, s.source, s.now)
	if err := s.store.Upsert(ctx, enriched); err != nil {
		return 0, err
	}
	return len(enriched), nil
}

// QueryByUser retrieves enriched posts for a specific user from the store.
func (s *Service) QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error) {
	return s.store.QueryByUser(ctx, userID)
}

func (s *Service) QueryRecent(ctx context.Context, limit, offset int) ([]models.EnrichedPost, error) {
	return s.store.QueryRecent(ctx, limit, offset)
}

func New(store StorePort, collector CollectorPort, source string, now func() time.Time) *Service {
	if now == nil {
		now = time.Now
	}
	return &Service{store: store, collector: collector, source: source, now: now}
}
