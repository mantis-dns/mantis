package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/mantis-dns/mantis/internal/config"
	mantisdns "github.com/mantis-dns/mantis/internal/dns"
	"github.com/mantis-dns/mantis/internal/domain"
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

	// Stub resolver until pipeline is wired (Task 9).
	resolver := &stubResolver{}

	dnsServer := mantisdns.NewServer(cfg.DNS.ListenAddress, resolver, logger)
	if err := dnsServer.Start(); err != nil {
		logger.Fatal().Err(err).Msg("failed to start DNS server")
	}
	logger.Info().Str("version", version).Msg("mantis started")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig

	logger.Info().Msg("shutting down")
	dnsServer.Stop()
}

// stubResolver returns SERVFAIL for all queries.
type stubResolver struct{}

func (s *stubResolver) Resolve(_ context.Context, _ *domain.Query) (*domain.Response, error) {
	return nil, fmt.Errorf("not implemented")
}
