package ingest

import (
	"testing"
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

func TestEnrich_Basic(t *testing.T) {
	// fixed time (non-UTC) to ensure UTC normalization is tested
	loc := time.FixedZone("IST", 5*60*60+30*60) // UTC+5:30
	fixedNow := func() time.Time {
		return time.Date(2025, 8, 17, 10, 11, 12, 0, loc)
	}
	exp := fixedNow().UTC()

	in := []models.Post{
		{UserID: 7, ID: 42, Title: "hello", Body: "world"},
	}
	out := Enrich(in, "placeholder_api", fixedNow)

	if len(out) != 1 {
		t.Fatalf("expected 1 item, got %d", len(out))
	}
	got := out[0]

	if got.UserID != 7 || got.ID != 42 {
		t.Errorf("unexpected ids: userID=%d id=%d", got.UserID, got.ID)
	}
	if got.Title != "hello" || got.Body != "world" {
		t.Errorf("unexpected content: title=%q body=%q", got.Title, got.Body)
	}
	if got.Source != "placeholder_api" {
		t.Errorf("expected source=placeholder_api, got %q", got.Source)
	}
	if !got.IngestedAt.Equal(exp) {
		t.Errorf("ingested_at not normalized to UTC. want=%s got=%s", exp, got.IngestedAt)
	}
	if got.IngestedAt.Location() != time.UTC {
		t.Errorf("ingested_at must be UTC, got %v", got.IngestedAt.Location())
	}
}

func TestEnrich_EmptyInput(t *testing.T) {
	fixedNow := func() time.Time { return time.Unix(1_720_000_000, 0) }
	out := Enrich(nil, "src", fixedNow)
	if len(out) != 0 {
		t.Fatalf("expected 0 items, got %d", len(out))
	}
}

func TestEnrich_DoesNotMutateInput(t *testing.T) {
	fixedNow := func() time.Time { return time.Unix(1_720_000_000, 0) }
	in := []models.Post{{UserID: 1, ID: 1, Title: "T", Body: "B"}}

	out := Enrich(in, "src", fixedNow)

	// mutate output and ensure input is unchanged
	out[0].Title = "CHANGED"
	if in[0].Title != "T" {
		t.Fatalf("input mutated; want %q got %q", "T", in[0].Title)
	}
}

func TestEnrich_MultipleItems(t *testing.T) {
	fixedNow := func() time.Time { return time.Unix(1_800_000_000, 0).In(time.FixedZone("X", 3*3600)) }
	exp := fixedNow().UTC()

	in := []models.Post{
		{UserID: 1, ID: 10, Title: "a", Body: "b"},
		{UserID: 1, ID: 11, Title: "c", Body: "d"},
	}
	out := Enrich(in, "srcX", fixedNow)

	if len(out) != 2 {
		t.Fatalf("expected 2 items, got %d", len(out))
	}
	for i, got := range out {
		if got.Source != "srcX" {
			t.Errorf("[%d] source mismatch: %q", i, got.Source)
		}
		if !got.IngestedAt.Equal(exp) || got.IngestedAt.Location() != time.UTC {
			t.Errorf("[%d] ingested_at UTC mismatch: want=%s got=%s loc=%v", i, exp, got.IngestedAt, got.IngestedAt.Location())
		}
		if got.UserID != in[i].UserID || got.ID != in[i].ID || got.Title != in[i].Title || got.Body != in[i].Body {
			t.Errorf("[%d] field mismatch: in=%+v out=%+v", i, in[i], got)
		}
	}
}

// Optional: lightweight benchmark to keep an eye on allocations/perf.
func BenchmarkEnrich(b *testing.B) {
	now := func() time.Time { return time.Unix(1_700_000_000, 0) }
	posts := make([]models.Post, 0, 1000)
	for i := 0; i < 1000; i++ {
		posts = append(posts, models.Post{UserID: i % 10, ID: i, Title: "t", Body: "b"})
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Enrich(posts, "bench", now)
	}
}
