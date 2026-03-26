package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"maxapp/internal/service"
)

// initDataHeader is passed by frontend for each request.
const initDataHeader = "X-Max-InitData"

// Auth validates MAX initData and injects user into context.
func Auth(authService *service.AuthService, logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// initData is sent via header (same string as WebApp.InitData).
			initData := r.Header.Get(initDataHeader)
			if initData == "" {
				writeAuthError(w, http.StatusUnauthorized, "missing initData")
				return
			}

			// Validate initData and resolve user record.
			user, err := authService.ValidateInitData(r.Context(), initData)
			if err != nil {
				logger.Warn("auth failed", "error", err)
				writeAuthError(w, http.StatusUnauthorized, "invalid initData")
				return
			}

			// Store user into context for downstream handlers.
			ctx := WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireApproved blocks unapproved or blocked users.
func RequireApproved() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.IsBlocked || !user.IsApproved {
				writeAuthError(w, http.StatusForbidden, "access denied")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAdmin allows only admin users.
func RequireAdmin() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())
			if !ok || user.IsBlocked || !user.IsAdmin {
				writeAuthError(w, http.StatusForbidden, "admin access required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// writeAuthError writes standardized auth error response.
func writeAuthError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"success": false,
		"message": message,
	})
}
