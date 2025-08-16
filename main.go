package main

import (
	"context"
	"log"
	"time"

	"github.com/renix-codex/ingestor/internal/api"
	"github.com/renix-codex/ingestor/internal/config"
	"github.com/renix-codex/ingestor/internal/ingest"
	"github.com/renix-codex/ingestor/internal/ingest/store"
	http "github.com/renix-codex/ingestor/internal/server"
)

func main() {
	cfg := config.FromEnv()
	ctx := context.Background()

	// adapters
	pg, err := store.New(ctx, cfg.BuildDSN())
	if err != nil {
		log.Fatalf("postgres init: %v", err)
	}
	col := ingest.NewHTTPCollector(cfg.SourceURL, cfg.HTTPTimeout)

	// service
	svc := ingest.New(pg, col, cfg.SourceName, time.Now)

	// api facade
	app := api.New(svc)

	// optional: one-shot ingest at startup via API (not directly via svc)
	ingCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	if n, err := app.IngestOnce(ingCtx); err != nil {
		log.Printf("ingest failed: %v", err)
	} else {
		log.Printf("ingested %d records", n)
	}
	cancel()

	// http server uses the api layer
	s := http.New(app)
	log.Printf("listening on %s", cfg.ListenAddr)
	if err := s.ListenAndServe(context.Background(), cfg.ListenAddr); err != nil {
		log.Fatal(err)
	}
}
