package handler

import "net/http"

// Health returns simple liveness indicator.
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  "ok",
	})
}

// Ready returns readiness indicator.
func (h *Handler) Ready(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"status":  "ready",
	})
}
