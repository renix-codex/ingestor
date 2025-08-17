package ingest

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

// Assumes:
//   func NewHTTPCollector(sourceURL string, timeout time.Duration) *HTTPCollector
//   func (c *HTTPCollector) Fetch(ctx context.Context) ([]models.Post, error)

func TestCollector_FetchOK(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"userId":1,"id":1,"title":"a","body":"b"}]`))
	}))
	defer s.Close()

	c := NewHTTPCollector(s.URL, 2*time.Second)
	posts, err := c.Fetch(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
}

func TestCollector_Timeout(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(750 * time.Millisecond) // exceed client timeout
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer s.Close()

	c := NewHTTPCollector(s.URL, 200*time.Millisecond)
	if _, err := c.Fetch(context.Background()); err == nil {
		t.Fatalf("expected timeout error, got nil")
	}
}

func TestCollector_InvalidStatus(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "upstream err", http.StatusBadGateway)
	}))
	defer s.Close()

	c := NewHTTPCollector(s.URL, 2*time.Second)
	if _, err := c.Fetch(context.Background()); err == nil {
		t.Fatalf("expected non-2xx error, got nil")
	}
}

func TestCollector_InvalidJSON(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"not":"an array"}`)) // your code expects an array of posts
	}))
	defer s.Close()

	c := NewHTTPCollector(s.URL, 2*time.Second)
	if _, err := c.Fetch(context.Background()); err == nil {
		t.Fatalf("expected JSON decode error, got nil")
	}
	_ = models.Post{} // keep import if your Fetch signature returns []models.Post
}
