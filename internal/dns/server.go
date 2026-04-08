package dns

import (
	"context"
	"fmt"
	"sync"

	mdns "codeberg.org/miekg/dns"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Server listens for DNS queries on UDP and TCP.
type Server struct {
	udpServer *mdns.Server
	tcpServer *mdns.Server
	resolver  domain.Resolver
	logger    zerolog.Logger
	wg        sync.WaitGroup
}

// NewServer creates a DNS server bound to the given address.
func NewServer(addr string, resolver domain.Resolver, logger zerolog.Logger) *Server {
	s := &Server{
		resolver: resolver,
		logger:   logger.With().Str("component", "dns").Logger(),
	}

	handler := &Handler{resolver: resolver, logger: s.logger}

	s.udpServer = &mdns.Server{
		Addr:    addr,
		Net:     "udp",
		Handler: handler,
	}
	s.tcpServer = &mdns.Server{
		Addr:    addr,
		Net:     "tcp",
		Handler: handler,
	}

	return s
}

// Start begins listening on UDP and TCP.
func (s *Server) Start() error {
	errCh := make(chan error, 2)

	s.wg.Add(2)
	go func() {
		defer s.wg.Done()
		s.logger.Info().Str("addr", s.udpServer.Addr).Msg("starting DNS server (UDP)")
		if err := s.udpServer.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("udp: %w", err)
		}
	}()
	go func() {
		defer s.wg.Done()
		s.logger.Info().Str("addr", s.tcpServer.Addr).Msg("starting DNS server (TCP)")
		if err := s.tcpServer.ListenAndServe(); err != nil {
			errCh <- fmt.Errorf("tcp: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// Stop gracefully shuts down the DNS server.
func (s *Server) Stop() {
	ctx := context.Background()
	s.udpServer.Shutdown(ctx)
	s.tcpServer.Shutdown(ctx)
	s.wg.Wait()
}
