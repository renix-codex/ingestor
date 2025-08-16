package ingest

import (
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

func Enrich(posts []models.Post, source string, now func() time.Time) []models.EnrichedPost {
	out := make([]models.EnrichedPost, 0, len(posts))
	for _, p := range posts {
		out = append(out, models.EnrichedPost{
			UserID:     p.UserID,
			ID:         p.ID,
			Title:      p.Title,
			Body:       p.Body,
			IngestedAt: now().UTC(),
			Source:     source,
		})
	}
	return out
}
