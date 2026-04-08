package dot

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"sync"

	mdns "codeberg.org/miekg/dns"
	"codeberg.org/miekg/dns/dnsutil"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Server implements DNS-over-TLS (RFC 7858).
type Server struct {
	listener net.Listener
	resolver domain.Resolver
	logger   zerolog.Logger
	wg       sync.WaitGroup
	done     chan struct{}
}

// NewServer creates a DoT server.
func NewServer(addr string, resolver domain.Resolver, tlsCfg *tls.Config, logger zerolog.Logger) (*Server, error) {
	ln, err := tls.Listen("tcp", addr, tlsCfg)
	if err != nil {
		return nil, err
	}
	return &Server{
		listener: ln,
		resolver: resolver,
		logger:   logger.With().Str("component", "dot").Logger(),
		done:     make(chan struct{}),
	}, nil
}

// Start begins accepting DoT connections.
func (s *Server) Start() {
	s.logger.Info().Str("addr", s.listener.Addr().String()).Msg("starting DoT server")
	go s.acceptLoop()
}

// Stop gracefully shuts down the DoT server.
func (s *Server) Stop() error {
	close(s.done)
	err := s.listener.Close()
	s.wg.Wait()
	return err
}

func (s *Server) acceptLoop() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.done:
				return
			default:
				s.logger.Error().Err(err).Msg("accept error")
				continue
			}
		}
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			s.handleConn(conn)
		}()
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	for {
		// Read 2-byte length prefix (RFC 7858).
		var length uint16
		if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
			if err != io.EOF {
				s.logger.Debug().Err(err).Msg("read length prefix")
			}
			return
		}

		buf := make([]byte, length)
		if _, err := io.ReadFull(conn, buf); err != nil {
			s.logger.Debug().Err(err).Msg("read message")
			return
		}

		msg := &mdns.Msg{Data: buf}
		if err := msg.Unpack(); err != nil {
			s.logger.Debug().Err(err).Msg("unpack message")
			return
		}

		if len(msg.Question) == 0 {
			continue
		}

		qName, qType := dnsutil.Question(msg)
		clientIP := conn.RemoteAddr().(*net.TCPAddr).IP

		q := &domain.Query{
			Domain:    dnsutil.Fqdn(qName),
			Type:      qType,
			ClientIP:  clientIP,
			Transport: domain.TransportDoT,
		}

		resp, err := s.resolver.Resolve(context.Background(), q)

		reply := new(mdns.Msg)
		dnsutil.SetReply(reply, msg)
		reply.RecursionAvailable = true

		if err != nil {
			reply.Rcode = mdns.RcodeServerFailure
		} else {
			for _, rr := range resp.Answers {
				dnsRR := domainRRtoDNS(rr, msg.Question[0].Header().Name)
				if dnsRR != nil {
					reply.Answer = append(reply.Answer, dnsRR)
				}
			}
		}

		reply.Pack()

		// Write 2-byte length prefix + message.
		var respLen [2]byte
		binary.BigEndian.PutUint16(respLen[:], uint16(len(reply.Data)))
		conn.Write(respLen[:])
		conn.Write(reply.Data)
	}
}
