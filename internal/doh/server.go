package doh

import (
	"context"
	"crypto/tls"
	"net/http"
	"time"

	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Server implements DNS-over-HTTPS (RFC 8484).
type Server struct {
	httpServer *http.Server
	logger     zerolog.Logger
}

// NewServer creates a DoH server.
func NewServer(addr string, resolver domain.Resolver, tlsCfg *tls.Config, logger zerolog.Logger) *Server {
	handler := &Handler{resolver: resolver, logger: logger}

	mux := http.NewServeMux()
	mux.Handle("/dns-query", handler)

	return &Server{
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			TLSConfig:         tlsCfg,
			ReadHeaderTimeout: 5 * time.Second,
			ReadTimeout:       10 * time.Second,
			WriteTimeout:      10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		logger: logger.With().Str("component", "doh").Logger(),
	}
}

// Start begins listening for DoH requests.
func (s *Server) Start() error {
	s.logger.Info().Str("addr", s.httpServer.Addr).Msg("starting DoH server")
	go func() {
		var err error
		if s.httpServer.TLSConfig != nil {
			err = s.httpServer.ListenAndServeTLS("", "")
		} else {
			err = s.httpServer.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			s.logger.Error().Err(err).Msg("DoH server error")
		}
	}()
	return nil
}

// Stop gracefully shuts down the DoH server.
func (s *Server) Stop() error {
	return s.httpServer.Shutdown(context.Background())
}

// Addr returns the listen address. Useful after Start with :0 port.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

