package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"github.com/go-chi/chi/v5"
	"go.opentelemetry.io/otel"
)

// adminUpdateRequest contains optional flag updates.
type adminUpdateRequest struct {
	IsApproved *bool `json:"isApproved"`
	IsBlocked  *bool `json:"isBlocked"`
	IsAdmin    *bool `json:"isAdmin"`
}

// ListUsers returns all users for admin view.
func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.admin")
	ctx, span := tracer.Start(r.Context(), "ListUsers")
	defer span.End()

	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		writeJSON(w, http.StatusOK, map[string]any{
			"success": true,
			"data":    []domain.User{},
		})
		return
	}

	// Fetch users by name.
	users, err := h.userService.SearchByName(ctx, query)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch users")
		return
	}

	// Return response.
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data":    users,
	})
}

// UpdateUserFlags changes approval/admin/block flags.
func (h *Handler) UpdateUserFlags(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.admin")
	ctx, span := tracer.Start(r.Context(), "UpdateUserFlags")
	defer span.End()

	// Validate id path param.
	idParam := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	// Parse payload.
	var req adminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// At least one field must be provided.
	if req.IsAdmin == nil && req.IsApproved == nil && req.IsBlocked == nil {
		writeError(w, http.StatusBadRequest, "nothing to update")
		return
	}

	// Update flags.
	if err := h.userService.UpdateFlags(ctx, userID, req.IsApproved, req.IsBlocked, req.IsAdmin); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	// Return success.
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}
