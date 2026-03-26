package handler

import (
	"encoding/json"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// validateRequest describes /auth/validate input.
type validateRequest struct {
	InitData string `json:"initData"`
}

// ValidateInitData validates initData and returns user profile flags.
func (h *Handler) ValidateInitData(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.auth")
	ctx, span := tracer.Start(r.Context(), "ValidateInitData")
	defer span.End()

	// Decode request body.
	var req validateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request")
		return
	}

	// Validate initData and fetch/create user.
	user, err := h.authService.ValidateInitData(ctx, req.InitData)
	if err != nil {
		span.SetAttributes(attribute.String("auth.error", err.Error()))
		writeError(w, http.StatusUnauthorized, "invalid initData")
		return
	}

	// Respond with minimal user flags for UI.
	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"userId":          user.ID,
			"isAdmin":         user.IsAdmin,
			"isApproved":      user.IsApproved,
			"isBlocked":       user.IsBlocked,
			"profileComplete": user.ProfileComplete(),
			"firstName":       user.FirstName,
			"lastName":        user.LastName,
		},
	})
}
