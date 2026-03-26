// Package handler defines HTTP handlers for the API.
package handler

import (
	"log/slog"

	"maxapp/internal/service"
)

// Handler aggregates all services and shared dependencies.
type Handler struct {
	authService      *service.AuthService
	userService      *service.UserService
	warehouseService *service.WarehouseService
	logger           *slog.Logger
	allowMockInitData bool
}

// NewHandler constructs Handler with all services.
func NewHandler(
	authService *service.AuthService,
	userService *service.UserService,
	warehouseService *service.WarehouseService,
	logger *slog.Logger,
	allowMockInitData bool,
) *Handler {
	return &Handler{
		authService:      authService,
		userService:      userService,
		warehouseService: warehouseService,
		logger:           logger,
		allowMockInitData: allowMockInitData,
	}
}
