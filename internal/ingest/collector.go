package ingest

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/renix-codex/ingestor/internal/models"
)

type Collector interface {
	Fetch(ctx context.Context) ([]models.Post, error)
}

type HTTPCollector struct {
	Client    *http.Client
	SourceURL string
}

func NewHTTPCollector(sourceURL string, timeout time.Duration) *HTTPCollector {
	return &HTTPCollector{
		SourceURL: sourceURL,
		Client: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   5 * time.Second,
					KeepAlive: 30 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout: 5 * time.Second,
				MaxIdleConns:        100,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *HTTPCollector) Fetch(ctx context.Context) ([]models.Post, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.SourceURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, errors.New("upstream returned non-2xx")
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var posts []models.Post
	if err := json.Unmarshal(body, &posts); err != nil {
		return nil, err
	}
	return posts, nil
}
