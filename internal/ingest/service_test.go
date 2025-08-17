package ingest

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

// Assumes you have:
//   type Service struct { ... }
//   func New(store StorePort, collector CollectorPort, source string, now func() time.Time) *Service
// and interfaces in this package:
//   type CollectorPort interface { Fetch(ctx context.Context) ([]models.Post, error) }
//   type StorePort interface {
//       Upsert(ctx context.Context, items []models.EnrichedPost) error
//       QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error)
//   }

type fakeCollectorOK struct{ items []models.Post }

func (f fakeCollectorOK) Fetch(ctx context.Context) ([]models.Post, error) { return f.items, nil }

type fakeCollectorErr struct{}

func (fakeCollectorErr) Fetch(ctx context.Context) ([]models.Post, error) {
	return nil, errors.New("upstream down")
}

type fakeStoreOK struct{ saved int }

func (f *fakeStoreOK) Upsert(ctx context.Context, items []models.EnrichedPost) error {
	f.saved = len(items)
	return nil
}
func (f *fakeStoreOK) QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error) {
	return []models.EnrichedPost{{
		UserID: userID, ID: 99, Title: "t", Body: "b",
		IngestedAt: time.Now().UTC(), Source: "src",
	}}, nil
}

// implement the exact signature expected by StorePort
func (f *fakeStoreOK) QueryRecent(ctx context.Context, limit, offset int) ([]models.EnrichedPost, error) {
	return []models.EnrichedPost{}, nil
}

type fakeStoreFail struct{}

func (fakeStoreFail) Upsert(ctx context.Context, items []models.EnrichedPost) error {
	return errors.New("db write failed")
}
func (fakeStoreFail) QueryByUser(ctx context.Context, userID int) ([]models.EnrichedPost, error) {
	return nil, errors.New("db read failed")
}

// implement the exact signature expected by StorePort
func (fakeStoreFail) QueryRecent(ctx context.Context, limit, offset int) ([]models.EnrichedPost, error) {
	return nil, errors.New("db recent read failed")
}

func TestService_IngestOnce_Success(t *testing.T) {
	store := &fakeStoreOK{}
	col := fakeCollectorOK{items: []models.Post{{UserID: 1, ID: 1, Title: "T", Body: "B"}}}
	svc := New(store, col, "src", func() time.Time { return time.Unix(1_720_000_000, 0) })

	n, err := svc.IngestOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 || store.saved != 1 {
		t.Fatalf("expected 1 saved item, got n=%d saved=%d", n, store.saved)
	}
}

func TestService_IngestOnce_DBError(t *testing.T) {
	store := fakeStoreFail{}
	col := fakeCollectorOK{items: []models.Post{{UserID: 1, ID: 2, Title: "T", Body: "B"}}}
	svc := New(store, col, "src", time.Now)

	if _, err := svc.IngestOnce(context.Background()); err == nil {
		t.Fatalf("expected db error, got nil")
	}
}

func TestService_IngestOnce_CollectorError(t *testing.T) {
	store := &fakeStoreOK{}
	col := fakeCollectorErr{}
	svc := New(store, col, "src", time.Now)

	if _, err := svc.IngestOnce(context.Background()); err == nil {
		t.Fatalf("expected collector error, got nil")
	}
}

func TestService_QueryByUser_DBError(t *testing.T) {
	store := fakeStoreFail{}
	col := fakeCollectorOK{}
	svc := New(store, col, "src", time.Now)

	if _, err := svc.QueryByUser(context.Background(), 1); err == nil {
		t.Fatalf("expected db read error, got nil")
	}
}
