package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mantis-dns/mantis/internal/config"
	mantisdns "github.com/mantis-dns/mantis/internal/dns"
	"github.com/mantis-dns/mantis/internal/gravity"
	"github.com/mantis-dns/mantis/internal/pipeline"
	"github.com/mantis-dns/mantis/internal/resolver"
	"github.com/rs/zerolog"
)

var version = "dev"

func main() {
	showVersion := flag.Bool("version", false, "Print version and exit")
	configPath := flag.String("config", "", "Path to configuration file")
	_ = flag.String("data-dir", "/var/lib/mantis", "Path to data directory")
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

	// Build components.
	dnsCache := resolver.NewDNSCache(cfg.DNS.CacheSize)
	forwarder := resolver.NewForwarder(cfg.DNS.Upstreams, logger)
	gravityEngine := gravity.NewEngine()

	// Build pipeline: cache -> gravity -> client rules -> upstream.
	chain := pipeline.NewChain(
		pipeline.NewCacheHandler(dnsCache),
		pipeline.NewGravityHandler(gravityEngine),
		pipeline.NewClientRuleHandler(),
		pipeline.NewUpstreamHandler(forwarder),
	)

	dnsServer := mantisdns.NewServer(cfg.DNS.ListenAddress, chain, logger)
	if err := dnsServer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start DNS server")
	}
	logger.Info().
		Str("version", version).
		Str("dns", cfg.DNS.ListenAddress).
		Int("upstreams", len(cfg.DNS.Upstreams)).
		Msg("mantis started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info().Msg("shutting down")
	dnsServer.Stop()
	dnsCache.Close()
}
