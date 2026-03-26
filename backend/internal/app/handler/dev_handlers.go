package handler

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/service"

	"go.opentelemetry.io/otel"
)

// DevInitData generates signed initData for local testing (dev-only).
func (h *Handler) DevInitData(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.dev")
	ctx, span := tracer.Start(r.Context(), "DevInitData")
	defer span.End()

	// Do not expose this endpoint unless explicitly enabled.
	if !h.allowMockInitData {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	// Parse query params.
	userID := int64(0)
	if rawID := r.URL.Query().Get("user_id"); rawID != "" {
		parsed, err := strconv.ParseInt(rawID, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid user_id")
			return
		}
		userID = parsed
	}

	input := service.InitDataInput{
		UserID:       userID,
		FirstName:    r.URL.Query().Get("first_name"),
		LastName:     r.URL.Query().Get("last_name"),
		Username:     r.URL.Query().Get("username"),
		LanguageCode: r.URL.Query().Get("language"),
		PhotoURL:     r.URL.Query().Get("photo_url"),
	}

	initData, err := h.authService.GenerateInitData(ctx, input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to generate initData")
		return
	}

	// Build optional link for convenience.
	base := r.URL.Query().Get("front")
	if base == "" {
		base = "http://localhost:5173"
	}
	base = strings.TrimRight(base, "/")
	// Encode initData as single query param.
	url := fmt.Sprintf("%s/?initData=%s", base, url.QueryEscape(initData))

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"data": map[string]any{
			"initData": initData,
			"url":      url,
		},
	})
}
