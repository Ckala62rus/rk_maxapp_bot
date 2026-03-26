package middleware

import (
	"net/http"

	"github.com/google/uuid"
)

// requestIDHeader is used to propagate request ID.
const requestIDHeader = "X-Request-Id"

// RequestID adds or generates request ID for every request.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prefer incoming request ID for tracing in upstream systems.
		requestID := r.Header.Get(requestIDHeader)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in context and send back in response.
		ctx := WithRequestID(r.Context(), requestID)
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
