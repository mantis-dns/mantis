package gravity

import (
	"context"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// SourceProvider returns the current list of blocklist sources and custom rules.
type SourceProvider interface {
	BlocklistSources(ctx context.Context) ([]domain.BlocklistSource, error)
	CustomRules(ctx context.Context) ([]domain.CustomRule, error)
}

// Scheduler runs periodic Gravity rebuilds.
type Scheduler struct {
	engine     *Engine
	downloader *Downloader
	provider   SourceProvider
	interval   time.Duration
	logger     zerolog.Logger
	cancel     context.CancelFunc
	done       chan struct{}
}

// NewScheduler creates a gravity rebuild scheduler.
func NewScheduler(engine *Engine, downloader *Downloader, provider SourceProvider, interval time.Duration, logger zerolog.Logger) *Scheduler {
	return &Scheduler{
		engine:     engine,
		downloader: downloader,
		provider:   provider,
		interval:   interval,
		logger:     logger.With().Str("component", "gravity-scheduler").Logger(),
		done:       make(chan struct{}),
	}
}

// Start begins the periodic rebuild loop. Runs an initial rebuild immediately.
func (s *Scheduler) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	go func() {
		defer close(s.done)

		// Initial rebuild.
		s.rebuild(ctx)

		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.rebuild(ctx)
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Stop cancels the scheduler and waits for completion.
func (s *Scheduler) Stop() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
}

// TriggerRebuild runs a one-off rebuild outside the schedule.
func (s *Scheduler) TriggerRebuild() {
	go s.rebuild(context.Background())
}

func (s *Scheduler) rebuild(ctx context.Context) {
	sources, err := s.provider.BlocklistSources(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to load blocklist sources")
		return
	}

	rules, err := s.provider.CustomRules(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("failed to load custom rules")
		return
	}

	s.engine.Rebuild(ctx, sources, rules, s.downloader, s.logger)
}
