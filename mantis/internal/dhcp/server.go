package dhcp

import (
	"context"
	"net"
	"time"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"github.com/insomniacslk/dhcp/dhcpv4/server4"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/rs/zerolog"
)

// Server implements a DHCPv4 server.
type Server struct {
	server    *server4.Server
	pool      *Pool
	leaseRepo domain.LeaseRepository
	dnsServer net.IP
	gateway   net.IP
	subnet    net.IPMask
	leaseDur  time.Duration
	logger    zerolog.Logger
}

// Config holds DHCP server configuration.
type Config struct {
	Interface    string
	RangeStart   string
	RangeEnd     string
	Gateway      string
	SubnetMask   string
	DNSServer    string
	LeaseDuration time.Duration
}

// NewServer creates a DHCP server.
func NewServer(cfg Config, leaseRepo domain.LeaseRepository, logger zerolog.Logger) (*Server, error) {
	pool, err := NewPool(cfg.RangeStart, cfg.RangeEnd)
	if err != nil {
		return nil, err
	}

	s := &Server{
		pool:      pool,
		leaseRepo: leaseRepo,
		dnsServer: net.ParseIP(cfg.DNSServer).To4(),
		gateway:   net.ParseIP(cfg.Gateway).To4(),
		subnet:    parseSubnet(cfg.SubnetMask),
		leaseDur:  cfg.LeaseDuration,
		logger:    logger.With().Str("component", "dhcp").Logger(),
	}

	// Load existing leases into pool.
	leases, err := leaseRepo.List(context.Background())
	if err == nil {
		for _, l := range leases {
			if ip := net.ParseIP(l.IP).To4(); ip != nil {
				pool.Reserve(ip, l.MAC)
			}
		}
	}

	addr := &net.UDPAddr{IP: net.IPv4zero, Port: 67}
	srv, err := server4.NewServer(cfg.Interface, addr, s.handleDHCP)
	if err != nil {
		return nil, err
	}
	s.server = srv
	return s, nil
}

// Start begins serving DHCP requests.
func (s *Server) Start() {
	s.logger.Info().Msg("starting DHCP server")
	go s.server.Serve()
}

// Stop stops the DHCP server.
func (s *Server) Stop() {
	s.server.Close()
}

func (s *Server) handleDHCP(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	switch msg.MessageType() {
	case dhcpv4.MessageTypeDiscover:
		s.handleDiscover(conn, peer, msg)
	case dhcpv4.MessageTypeRequest:
		s.handleRequest(conn, peer, msg)
	case dhcpv4.MessageTypeRelease:
		s.handleRelease(msg)
	}
}

func (s *Server) handleDiscover(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	mac := msg.ClientHWAddr.String()
	ip, err := s.pool.Allocate(mac)
	if err != nil {
		s.logger.Warn().Err(err).Str("mac", mac).Msg("pool exhausted")
		return
	}

	resp, err := dhcpv4.NewReplyFromRequest(msg,
		dhcpv4.WithMessageType(dhcpv4.MessageTypeOffer),
		dhcpv4.WithYourIP(ip),
		dhcpv4.WithServerIP(s.dnsServer),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(s.subnet)),
		dhcpv4.WithOption(dhcpv4.OptRouter(s.gateway)),
		dhcpv4.WithOption(dhcpv4.OptDNS(s.dnsServer)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(s.leaseDur)),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("create OFFER")
		return
	}

	conn.WriteTo(resp.ToBytes(), peer)
	s.logger.Debug().Str("mac", mac).Str("ip", ip.String()).Msg("OFFER sent")
}

func (s *Server) handleRequest(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	mac := msg.ClientHWAddr.String()
	ip, err := s.pool.Allocate(mac)
	if err != nil {
		s.sendNAK(conn, peer, msg)
		return
	}

	hostname := msg.HostName()

	lease := &domain.DhcpLease{
		MAC:      mac,
		IP:       ip.String(),
		Hostname: hostname,
		LeaseEnd: time.Now().Add(s.leaseDur),
	}
	s.leaseRepo.Create(context.Background(), lease)

	resp, err := dhcpv4.NewReplyFromRequest(msg,
		dhcpv4.WithMessageType(dhcpv4.MessageTypeAck),
		dhcpv4.WithYourIP(ip),
		dhcpv4.WithServerIP(s.dnsServer),
		dhcpv4.WithOption(dhcpv4.OptSubnetMask(s.subnet)),
		dhcpv4.WithOption(dhcpv4.OptRouter(s.gateway)),
		dhcpv4.WithOption(dhcpv4.OptDNS(s.dnsServer)),
		dhcpv4.WithOption(dhcpv4.OptIPAddressLeaseTime(s.leaseDur)),
	)
	if err != nil {
		s.logger.Error().Err(err).Msg("create ACK")
		return
	}

	conn.WriteTo(resp.ToBytes(), peer)
	s.logger.Info().Str("mac", mac).Str("ip", ip.String()).Str("hostname", hostname).Msg("ACK sent")
}

func (s *Server) handleRelease(msg *dhcpv4.DHCPv4) {
	mac := msg.ClientHWAddr.String()
	if ip := msg.YourIPAddr; ip != nil {
		s.pool.Release(ip)
	}
	s.leaseRepo.Delete(context.Background(), mac)
	s.logger.Info().Str("mac", mac).Msg("RELEASE")
}

func (s *Server) sendNAK(conn net.PacketConn, peer net.Addr, msg *dhcpv4.DHCPv4) {
	resp, err := dhcpv4.NewReplyFromRequest(msg,
		dhcpv4.WithMessageType(dhcpv4.MessageTypeNak),
	)
	if err != nil {
		return
	}
	conn.WriteTo(resp.ToBytes(), peer)
}

func parseSubnet(mask string) net.IPMask {
	ip := net.ParseIP(mask).To4()
	if ip == nil {
		return net.CIDRMask(24, 32)
	}
	return net.IPv4Mask(ip[0], ip[1], ip[2], ip[3])
}
