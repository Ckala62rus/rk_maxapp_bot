package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/app/middleware"

	"go.opentelemetry.io/otel"
)

// profileRequest describes profile update payload.
type profileRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// UpdateProfile stores user first/last name.
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.user")
	ctx, span := tracer.Start(r.Context(), "UpdateProfile")
	defer span.End()

	// Ensure user is present in context.
	user, ok := middleware.UserFromContext(ctx)
	if !ok || user.IsBlocked {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	// Parse request body.
	var req profileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Validate payload.
	req.FirstName = strings.TrimSpace(req.FirstName)
	req.LastName = strings.TrimSpace(req.LastName)
	if req.FirstName == "" || req.LastName == "" {
		writeError(w, http.StatusBadRequest, "firstName and lastName required")
		return
	}

	// Save profile.
	if err := h.userService.UpdateProfile(ctx, user.ID, req.FirstName, req.LastName); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	// Return success.
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
	})
}
