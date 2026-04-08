package gravity

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// DownloadResult holds the result of downloading and parsing one blocklist.
type DownloadResult struct {
	Source  domain.BlocklistSource
	Domains []string
	Err     error
}

// Downloader fetches blocklists over HTTP.
type Downloader struct {
	client *http.Client
	logger zerolog.Logger
}

// NewDownloader creates a downloader with a 30s timeout.
func NewDownloader(logger zerolog.Logger) *Downloader {
	return &Downloader{
		client: &http.Client{Timeout: 30 * time.Second},
		logger: logger.With().Str("component", "downloader").Logger(),
	}
}

// DownloadAll fetches all enabled blocklist sources with max concurrency.
func (d *Downloader) DownloadAll(ctx context.Context, sources []domain.BlocklistSource) []DownloadResult {
	results := make(chan DownloadResult, len(sources))
	sem := make(chan struct{}, 5) // max 5 concurrent downloads

	for _, src := range sources {
		if !src.Enabled {
			continue
		}
		go func(s domain.BlocklistSource) {
			sem <- struct{}{}
			defer func() { <-sem }()
			domains, err := d.download(ctx, s)
			results <- DownloadResult{Source: s, Domains: domains, Err: err}
		}(src)
	}

	var enabled int
	for _, s := range sources {
		if s.Enabled {
			enabled++
		}
	}

	out := make([]DownloadResult, 0, enabled)
	for range enabled {
		r := <-results
		if r.Err != nil {
			d.logger.Warn().Err(r.Err).Str("source", r.Source.Name).Msg("download failed")
		} else {
			d.logger.Info().Str("source", r.Source.Name).Int("domains", len(r.Domains)).Msg("downloaded")
		}
		out = append(out, r)
	}
	return out
}

func (d *Downloader) download(ctx context.Context, src domain.BlocklistSource) ([]string, error) {
	var lastErr error
	for attempt := range 3 {
		domains, err := d.tryDownload(ctx, src)
		if err == nil {
			return domains, nil
		}
		lastErr = err
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		d.logger.Debug().Err(err).Str("source", src.Name).Int("attempt", attempt+1).Msg("retry")

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(backoff):
		}
	}
	return nil, fmt.Errorf("after 3 attempts: %w", lastErr)
}

func (d *Downloader) tryDownload(ctx context.Context, src domain.BlocklistSource) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Limit body to 100MB.
	limited := io.LimitReader(resp.Body, 100*1024*1024)
	result := Parse(limited, src.Format)
	return result.Domains, nil
}
