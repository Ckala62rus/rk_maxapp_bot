package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// statusRecorder captures HTTP status code for logging/metrics.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

// WriteHeader records status code before writing response.
func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// Logging adds structured request logs.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Start timer for duration calculation.
			start := time.Now()
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			// Execute handler.
			next.ServeHTTP(rec, r)

			// Extract context values for audit.
			requestID := RequestIDFromContext(r.Context())
			userID := int64(0)
			if user, ok := UserFromContext(r.Context()); ok {
				userID = user.ID
			}

			// Log basic request information.
			logger.Info("request completed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", rec.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", requestID,
				"user_id", userID,
			)
		})
	}
}
