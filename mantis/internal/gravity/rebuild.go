package gravity

import (
	"context"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// RebuildStatus holds the result of a Gravity rebuild.
type RebuildStatus struct {
	TotalDomains int           `json:"totalDomains"`
	Sources      int           `json:"sources"`
	Duration     time.Duration `json:"duration"`
	Timestamp    time.Time     `json:"timestamp"`
}

// Rebuild downloads all enabled blocklists, deduplicates, merges with custom rules,
// and atomically swaps the block tree.
func (e *Engine) Rebuild(ctx context.Context, sources []domain.BlocklistSource, customRules []domain.CustomRule, downloader *Downloader, logger zerolog.Logger) RebuildStatus {
	start := time.Now()

	results := downloader.DownloadAll(ctx, sources)

	// Deduplicate all domains.
	seen := make(map[string]struct{})
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		for _, d := range r.Domains {
			seen[d] = struct{}{}
		}
	}

	// Merge custom block rules.
	for _, rule := range customRules {
		if rule.Type == domain.RuleBlock {
			seen[rule.Domain] = struct{}{}
		}
	}

	domains := make([]string, 0, len(seen))
	for d := range seen {
		domains = append(domains, d)
	}

	e.RebuildFromDomains(domains)

	// Build allow tree from custom allow rules.
	var allowDomains []string
	for _, rule := range customRules {
		if rule.Type == domain.RuleAllow {
			allowDomains = append(allowDomains, rule.Domain)
		}
	}
	if len(allowDomains) > 0 {
		e.SetAllowRules(allowDomains)
	}

	status := RebuildStatus{
		TotalDomains: len(domains),
		Sources:      len(results),
		Duration:     time.Since(start),
		Timestamp:    time.Now(),
	}

	logger.Info().
		Int("domains", status.TotalDomains).
		Int("sources", status.Sources).
		Dur("duration", status.Duration).
		Msg("gravity rebuild complete")

	return status
}
