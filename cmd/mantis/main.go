package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mantis-dns/mantis/internal/api"
	"github.com/mantis-dns/mantis/internal/config"
	mantisdns "github.com/mantis-dns/mantis/internal/dns"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
	"github.com/mantis-dns/mantis/internal/gravity"
	"github.com/mantis-dns/mantis/internal/pipeline"
	"github.com/mantis-dns/mantis/internal/resolver"
	"github.com/mantis-dns/mantis/internal/stats"
	"github.com/mantis-dns/mantis/internal/storage"
	"github.com/mantis-dns/mantis/internal/web"
	"github.com/rs/zerolog"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "", "Path to configuration file")
	dataDir := flag.String("data-dir", "/var/lib/mantis", "Path to data directory")
	flag.Parse()

	if *showVersion {
		fmt.Printf("mantis version %s\n", version)
		os.Exit(0)
	}

	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	if *dataDir != "" {
		cfg.Storage.DataDir = *dataDir
	}

	// Open storage.
	db, err := storage.Open(cfg.Storage.DataDir)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to open storage")
	}
	defer db.Close()

	pdb := db.Pebble()

	// Create repositories.
	queryLogStore := storage.NewQueryLogStore(pdb)
	defer queryLogStore.Close()
	blocklistStore := storage.NewBlocklistStore(pdb)
	ruleStore := storage.NewRuleStore(pdb)
	leaseStore := storage.NewLeaseStore(pdb)
	settingsStore := storage.NewSettingsStore(pdb)
	sessionStore := storage.NewSessionStore(pdb)

	// Build components.
	eventBus := event.NewBus()
	dnsCache := resolver.NewDNSCache(cfg.DNS.CacheSize)
	forwarder := resolver.NewForwarder(cfg.DNS.Upstreams, logger)
	gravityEngine := gravity.NewEngine()

	// Stats collector.
	collector := stats.NewCollector(pdb, eventBus)
	collector.Start()
	defer collector.Stop()

	aggregator := stats.NewAggregator(pdb, collector)

	// Query log writer subscribes to event bus.
	go func() {
		ch := eventBus.Subscribe(10000)
		for ev := range ch {
			entry := eventToLogEntry(ev)
			queryLogStore.Append(context.Background(), &entry)
		}
	}()

	// Build pipeline: cache -> gravity -> client rules -> upstream.
	chain := pipeline.NewChain(eventBus,
		pipeline.NewCacheHandler(dnsCache),
		pipeline.NewGravityHandler(gravityEngine),
		pipeline.NewClientRuleHandler(),
		pipeline.NewUpstreamHandler(forwarder),
	)

	// Start DNS server.
	dnsServer := mantisdns.NewServer(cfg.DNS.ListenAddress, chain, logger)
	if err := dnsServer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start DNS server")
	}

	// Build API router.
	router := api.NewRouter(&api.Dependencies{
		Sessions:   sessionStore,
		Settings:   settingsStore,
		Blocklists: blocklistStore,
		Rules:      ruleStore,
		QueryLog:   queryLogStore,
		Leases:     leaseStore,
		Gravity:    gravityEngine,
		Stats:      aggregator,
		EventBus:   eventBus,
		Logger:     logger,
		Version:    version,
		RateLimit:  cfg.API.RateLimit,
		APIHost:    cfg.API.ListenAddress,
	})

	// Mount SPA at root.
	router.Handle("/*", web.SPAHandler())

	// Start API/UI server.
	apiServer := &http.Server{
		Addr:    cfg.API.ListenAddress,
		Handler: router,
	}
	go func() {
		logger.Info().Str("addr", cfg.API.ListenAddress).Msg("starting API server")
		if err := apiServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error().Err(err).Msg("API server error")
		}
	}()

	logger.Info().
		Str("version", version).
		Str("dns", cfg.DNS.ListenAddress).
		Str("api", cfg.API.ListenAddress).
		Int("upstreams", len(cfg.DNS.Upstreams)).
		Msg("mantis started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info().Msg("shutting down")
	apiServer.Close()
	dnsServer.Stop()
	dnsCache.Close()
}

func eventToLogEntry(ev event.QueryEvent) domain.QueryLogEntry {
	return domain.QueryLogEntry{
		Timestamp: ev.Timestamp,
		ClientIP:  ev.ClientIP,
		Domain:    ev.Domain,
		QueryType: ev.QueryType,
		Result:    ev.Result,
		Upstream:  ev.Upstream,
		LatencyUs: ev.LatencyUs,
		Answer:    ev.Answer,
	}
}
