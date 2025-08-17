package store

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/renix-codex/ingestor/internal/ingest" // import ONLY for the interface
	"github.com/renix-codex/ingestor/internal/models" // import ONLY for the types
)

type PGStore struct{ pool *pgxpool.Pool }

// Ensure PGStore implements the ingest.StorePort interface.
var _ ingest.StorePort = (*PGStore)(nil)

func New(ctx context.Context, dsn string) (*PGStore, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	_, err = pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS posts (
  user_id INT NOT NULL,
  id INT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  ingested_at TIMESTAMPTZ NOT NULL,
  source TEXT NOT NULL,
  doc JSONB NOT NULL,
  PRIMARY KEY (user_id, id)
);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_doc_gin ON posts USING GIN (doc);
`)
	if err != nil {
		return nil, err
	}
	return &PGStore{pool: pool}, nil
}

// --- your exact methods, unchanged ---

func (s *PGStore) Upsert(ctx context.Context, items []models.EnrichedPost) error {
	b := &pgx.Batch{}
	for _, it := range items {
		raw, _ := json.Marshal(it)
		b.Queue(`
INSERT INTO posts (user_id,id,title,body,ingested_at,source,doc)
VALUES ($1,$2,$3,$4,$5,$6,$7)
ON CONFLICT (user_id,id) DO UPDATE SET
  title=EXCLUDED.title, body=EXCLUDED.body,
  ingested_at=EXCLUDED.ingested_at, source=EXCLUDED.source, doc=EXCLUDED.doc`,
			it.UserID, it.ID, it.Title, it.Body, it.IngestedAt, it.Source, raw)
	}
	br := s.pool.SendBatch(ctx, b)
	defer br.Close()
	for range items {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}
	return nil
}

func (s *PGStore) QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error) {
	rows, err := s.pool.Query(ctx, `SELECT doc FROM posts WHERE user_id=$1 ORDER BY id`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.EnrichedPost
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var e models.EnrichedPost
		if err := json.Unmarshal(raw, &e); err == nil {
			out = append(out, e)
		}
	}
	return out, rows.Err()
}

func (s *PGStore) QueryRecent(ctx context.Context, limit, offset int) ([]models.EnrichedPost, error) {
	// sane limits
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.pool.Query(ctx, `
SELECT doc
FROM posts
ORDER BY ingested_at DESC, id DESC
LIMIT $1 OFFSET $2
`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []models.EnrichedPost
	for rows.Next() {
		var raw []byte
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var e models.EnrichedPost
		if err := json.Unmarshal(raw, &e); err == nil {
			out = append(out, e)
		}
	}
	return out, rows.Err()
}
