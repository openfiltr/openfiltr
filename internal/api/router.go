package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/openfiltr/openfiltr/internal/auth"
	"github.com/openfiltr/openfiltr/internal/config"
	"github.com/openfiltr/openfiltr/internal/storage"
)

func NewRouter(cfg *config.Config, db storage.Store, version string) http.Handler {
	r := chi.NewRouter()
	authSvc := auth.NewService(db, cfg.Auth.JWTSecret, cfg.Auth.TokenExpiry)
	h := &Handler{db: db, cfg: cfg, authSvc: authSvc, version: version}

	r.Use(middleware.RequestID, middleware.RealIP, middleware.Logger, middleware.Recoverer, middleware.StripSlashes)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID", csrfHeaderName},
		ExposedHeaders:   []string{"Link", "X-Total-Count", csrfHeaderName},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public
	r.Post("/api/v1/auth/login", h.Login)
	r.Post("/api/v1/auth/setup", h.Setup)
	r.Get("/api/v1/system/health", h.Health)
	r.Get("/api/v1/system/version", h.Version)

	// Protected
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/api/v1/system/status", h.Status)
		r.Get("/api/v1/system/stats", h.Stats)

		r.Post("/api/v1/auth/logout", h.Logout)
		r.Get("/api/v1/auth/me", h.Me)
		r.Get("/api/v1/auth/tokens", h.ListTokens)
		r.Post("/api/v1/auth/tokens", h.CreateToken)
		r.Delete("/api/v1/auth/tokens/{id}", h.DeleteToken)

		r.Get("/api/v1/filtering/block-rules", h.ListBlockRules)
		r.Post("/api/v1/filtering/block-rules", h.CreateBlockRule)
		r.Get("/api/v1/filtering/block-rules/{id}", h.GetBlockRule)
		r.Put("/api/v1/filtering/block-rules/{id}", h.UpdateBlockRule)
		r.Delete("/api/v1/filtering/block-rules/{id}", h.DeleteBlockRule)

		r.Get("/api/v1/filtering/allow-rules", h.ListAllowRules)
		r.Post("/api/v1/filtering/allow-rules", h.CreateAllowRule)
		r.Get("/api/v1/filtering/allow-rules/{id}", h.GetAllowRule)
		r.Put("/api/v1/filtering/allow-rules/{id}", h.UpdateAllowRule)
		r.Delete("/api/v1/filtering/allow-rules/{id}", h.DeleteAllowRule)

		r.Get("/api/v1/filtering/sources", h.ListRuleSources)
		r.Post("/api/v1/filtering/sources", h.CreateRuleSource)
		r.Get("/api/v1/filtering/sources/{id}", h.GetRuleSource)
		r.Put("/api/v1/filtering/sources/{id}", h.UpdateRuleSource)
		r.Delete("/api/v1/filtering/sources/{id}", h.DeleteRuleSource)
		r.Post("/api/v1/filtering/sources/{id}/refresh", h.RefreshRuleSource)

		r.Get("/api/v1/dns/upstream-servers", h.ListUpstreamServers)
		r.Post("/api/v1/dns/upstream-servers", h.CreateUpstreamServer)
		r.Get("/api/v1/dns/upstream-servers/{id}", h.GetUpstreamServer)
		r.Put("/api/v1/dns/upstream-servers/{id}", h.UpdateUpstreamServer)
		r.Delete("/api/v1/dns/upstream-servers/{id}", h.DeleteUpstreamServer)

		r.Get("/api/v1/dns/entries", h.ListDNSEntries)
		r.Post("/api/v1/dns/entries", h.CreateDNSEntry)
		r.Get("/api/v1/dns/entries/{id}", h.GetDNSEntry)
		r.Put("/api/v1/dns/entries/{id}", h.UpdateDNSEntry)
		r.Delete("/api/v1/dns/entries/{id}", h.DeleteDNSEntry)

		r.Get("/api/v1/clients", h.ListClients)
		r.Post("/api/v1/clients", h.CreateClient)
		r.Get("/api/v1/clients/{id}", h.GetClient)
		r.Put("/api/v1/clients/{id}", h.UpdateClient)
		r.Delete("/api/v1/clients/{id}", h.DeleteClient)

		r.Get("/api/v1/groups", h.ListGroups)
		r.Post("/api/v1/groups", h.CreateGroup)
		r.Get("/api/v1/groups/{id}", h.GetGroup)
		r.Put("/api/v1/groups/{id}", h.UpdateGroup)
		r.Delete("/api/v1/groups/{id}", h.DeleteGroup)

		r.Get("/api/v1/activity", h.ListActivity)
		r.Get("/api/v1/activity/stats", h.ActivityStats)

		r.Get("/api/v1/config/export", h.ExportConfig)
		r.Post("/api/v1/config/import", h.ImportConfig)

		r.Get("/api/v1/audit", h.ListAuditEvents)
	})

	// Serve OpenAPI spec
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "openapi/openapi.yaml")
	})

	return r
}
