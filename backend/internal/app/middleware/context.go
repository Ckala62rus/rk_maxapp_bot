// Package middleware provides HTTP middleware and context helpers.
package middleware

import (
	"context"

	"maxapp/internal/domain"
)

// contextKey is a typed key for context values.
type contextKey string

// Context keys.
const (
	requestIDKey contextKey = "request_id"
	userKey      contextKey = "user"
)

// WithRequestID stores request ID in context.
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey, requestID)
}

// RequestIDFromContext returns request ID if present.
func RequestIDFromContext(ctx context.Context) string {
	value, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}
	return value
}

// WithUser stores user in context.
func WithUser(ctx context.Context, user *domain.User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// UserFromContext returns user from context.
func UserFromContext(ctx context.Context) (*domain.User, bool) {
	value, ok := ctx.Value(userKey).(*domain.User)
	return value, ok
}
