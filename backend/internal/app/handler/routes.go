package handler

import (
	"net/http"

	"maxapp/internal/app/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// InitRoutes wires middlewares and endpoints.
func (h *Handler) InitRoutes() http.Handler {
	r := chi.NewRouter()

	// Global middlewares.
	r.Use(middleware.RequestID)
	r.Use(middleware.Logging(h.logger))
	r.Use(middleware.Metrics())

	// Health endpoints for orchestrators.
	r.Get("/healthz", h.Health)
	r.Get("/readyz", h.Ready)
	r.Handle("/metrics", promhttp.Handler())

	// API routes.
	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/validate", h.ValidateInitData)
		// Dev-only endpoint to generate initData for testing.
		r.Get("/dev/init-data", h.DevInitData)

		// Routes that require only auth.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(h.authService, h.logger))

			// Profile update requires auth but not approval.
			r.Post("/users/profile", h.UpdateProfile)

			// Routes that require approved and not blocked user.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireApproved())
				r.Get("/warehouse/batches", h.SearchBatches)
			})

			// Admin routes.
			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireAdmin())
				r.Get("/admin/users", h.ListUsers)
				r.Patch("/admin/users/{id}", h.UpdateUserFlags)
			})
		})
	})

	return r
}
