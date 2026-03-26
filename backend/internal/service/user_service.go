package service

import (
	"context"
	"log/slog"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"go.opentelemetry.io/otel"
)

// UserService encapsulates user-related business logic.
type UserService struct {
	repo   domain.UserRepository
	logger *slog.Logger
}

// NewUserService constructs UserService.
func NewUserService(repo domain.UserRepository, logger *slog.Logger) *UserService {
	return &UserService{repo: repo, logger: logger}
}

// UpdateProfile saves first/last name for user.
func (s *UserService) UpdateProfile(ctx context.Context, userID int64, firstName, lastName string) error {
	tracer := otel.Tracer("service.user")
	ctx, span := tracer.Start(ctx, "UpdateProfile")
	defer span.End()

	// Log in debug to avoid leaking PII in prod info logs.
	s.logger.Debug("updating profile", "user_id", userID)
	return s.repo.UpdateProfile(ctx, userID, firstName, lastName)
}

// List returns all users for admin view.
func (s *UserService) List(ctx context.Context) ([]domain.User, error) {
	tracer := otel.Tracer("service.user")
	ctx, span := tracer.Start(ctx, "ListUsers")
	defer span.End()

	return s.repo.List(ctx)
}

// SearchByName finds users by profile first/last name.
func (s *UserService) SearchByName(ctx context.Context, query string) ([]domain.User, error) {
	tracer := otel.Tracer("service.user")
	ctx, span := tracer.Start(ctx, "SearchByName")
	defer span.End()

	return s.repo.SearchByName(ctx, query)
}

// UpdateFlags updates approval/admin/block flags.
func (s *UserService) UpdateFlags(ctx context.Context, id int64, isApproved, isBlocked, isAdmin *bool) error {
	tracer := otel.Tracer("service.user")
	ctx, span := tracer.Start(ctx, "UpdateFlags")
	defer span.End()

	// Info log for admin actions.
	s.logger.Info("admin updated flags", "user_id", id, "approved", isApproved, "blocked", isBlocked, "admin", isAdmin)
	return s.repo.UpdateFlags(ctx, id, isApproved, isBlocked, isAdmin)
}
