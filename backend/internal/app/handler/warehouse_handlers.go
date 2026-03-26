package handler

import (
	"errors"
	"net/http"
	"strings"

	"maxapp/internal/app/middleware"
	"maxapp/internal/domain"
	"maxapp/internal/service"

	"go.opentelemetry.io/otel"
)

// SearchBatches handles DAX search endpoint.
func (h *Handler) SearchBatches(w http.ResponseWriter, r *http.Request) {
	tracer := otel.Tracer("handler.warehouse")
	ctx, span := tracer.Start(r.Context(), "SearchBatches")
	defer span.End()

	// Ensure authenticated user exists in context.
	user, ok := middleware.UserFromContext(ctx)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Read code query param.
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	if code == "" {
		writeError(w, http.StatusBadRequest, "code is required")
		return
	}

	// Call service and return response.
	batches, err := h.warehouseService.SearchBatches(ctx, user.ID, code)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCode) {
			writeError(w, http.StatusBadRequest, "invalid code format")
			return
		}
		if errors.Is(err, domain.ErrWarehouseUnavailable) {
			writeError(w, http.StatusServiceUnavailable, "проблемы с соединением БД")
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to fetch batches")
		return
	}

	if len(batches) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]any{
			"success": false,
			"rows":    0,
			"data":    []any{},
			"message": "партия не найдена",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"success": true,
		"rows":    len(batches),
		"data":    batches,
	})
}
