package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mantis-dns/mantis/internal/domain"
	"github.com/mantis-dns/mantis/internal/event"
	"github.com/mantis-dns/mantis/internal/gravity"
	"github.com/mantis-dns/mantis/internal/stats"
	"github.com/rs/zerolog"
)

// Dependencies holds all API handler dependencies.
type Dependencies struct {
	Sessions   domain.SessionRepository
	Settings   domain.SettingsRepository
	Blocklists domain.BlocklistRepository
	Rules      domain.CustomRuleRepository
	QueryLog   domain.QueryLogRepository
	Leases     domain.LeaseRepository
	Gravity    *gravity.Engine
	Stats      *stats.Aggregator
	EventBus   *event.Bus
	Logger     zerolog.Logger
	Version    string
	RateLimit  int
	APIHost    string
}

// NewRouter creates the chi router with all API routes.
func NewRouter(deps *Dependencies) chi.Router {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(CORSMiddleware(deps.APIHost))
	r.Use(SecurityHeaders)

	auth := &AuthHandler{sessions: deps.Sessions, settings: deps.Settings}
	statsH := &StatsHandler{aggregator: deps.Stats}
	queries := &QueryHandler{queryLog: deps.QueryLog, bus: deps.EventBus}
	SetTrustedAPIHost(deps.APIHost)
	blocklists := &BlocklistHandler{repo: deps.Blocklists}
	rules := &RulesHandler{repo: deps.Rules, gravity: deps.Gravity}
	dhcpH := &DHCPHandler{leases: deps.Leases}
	settings := &SettingsHandler{repo: deps.Settings}
	system := &SystemHandler{version: deps.Version, gravity: deps.Gravity}

	r.Route("/api/v1", func(r chi.Router) {
		// Public routes.
		r.Post("/auth/setup", func(w http.ResponseWriter, r *http.Request) {
			RateLimitMiddleware(deps.RateLimit)(http.HandlerFunc(auth.Setup)).ServeHTTP(w, r)
		})
		r.Post("/auth/login", auth.Login)
		r.Get("/system/health", system.Health)

		// Protected routes.
		r.Group(func(r chi.Router) {
			r.Use(AuthMiddleware(deps.Sessions))
			r.Use(RateLimitMiddleware(deps.RateLimit))

			r.Post("/auth/logout", auth.Logout)

			r.Get("/stats/summary", statsH.Summary)
			r.Get("/stats/overtime", statsH.Overtime)
			r.Get("/stats/top-domains", statsH.TopDomains)
			r.Get("/stats/top-clients", statsH.TopClients)

			r.Get("/queries", queries.List)
			r.Get("/queries/stream", queries.Stream)

			r.Get("/blocklists", blocklists.List)
			r.Post("/blocklists", blocklists.Create)
			r.Put("/blocklists/{id}", blocklists.Update)
			r.Delete("/blocklists/{id}", blocklists.Delete)

			r.Post("/gravity/rebuild", system.RebuildGravity)
			r.Get("/gravity/status", system.GravityStatus)

			r.Get("/rules", rules.List)
			r.Post("/rules", rules.Create)
			r.Delete("/rules/{id}", rules.Delete)

			r.Get("/dhcp/leases", dhcpH.ListLeases)
			r.Post("/dhcp/leases/static", dhcpH.CreateStaticLease)
			r.Delete("/dhcp/leases/static/{mac}", dhcpH.DeleteStaticLease)

			r.Get("/settings", settings.GetAll)
			r.Put("/settings", settings.Update)

			r.Get("/system/info", system.Info)
			r.Post("/system/restart-dns", system.RestartDNS)
		})
	})

	return r
}
