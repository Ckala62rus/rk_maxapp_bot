package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// ErrInvalidCode indicates code does not match allowed formats.
var ErrInvalidCode = errors.New("invalid code format")

// WarehouseService encapsulates DAX search logic and history.
type WarehouseService struct {
	repo        domain.WarehouseRepository
	historyRepo domain.SearchHistoryRepository
	logger      *slog.Logger
}

// NewWarehouseService constructs WarehouseService.
func NewWarehouseService(repo domain.WarehouseRepository, historyRepo domain.SearchHistoryRepository, logger *slog.Logger) *WarehouseService {
	return &WarehouseService{
		repo:        repo,
		historyRepo: historyRepo,
		logger:      logger,
	}
}

// SearchBatches validates code, queries DAX and stores history.
func (s *WarehouseService) SearchBatches(ctx context.Context, userID int64, code string) ([]domain.Batch, error) {
	tracer := otel.Tracer("service.warehouse")
	ctx, span := tracer.Start(ctx, "SearchBatches")
	defer span.End()

	// Validate and normalize code format.
	parsed, mode, err := parseCode(code)
	if err != nil {
		return nil, ErrInvalidCode
	}

	span.SetAttributes(attribute.String("code", parsed), attribute.Int64("user_id", userID))
	s.logger.Info("warehouse search start", "code", code, "normalized", parsed, "mode", mode, "user_id", userID)
	// Measure query duration for audit.
	start := time.Now()

	batches, err := s.repo.FindBatches(ctx, parsed, mode)
	durationMs := time.Since(start).Milliseconds()

	// Log all results for debugging (requested).
	if err == nil {
		s.logger.Info(
			"warehouse search results",
			"code", code,
			"normalized", parsed,
			"mode", mode,
			"rows", len(batches),
			"results", batches,
		)
	}

	// Persist search history regardless of query result.
	history := domain.SearchHistory{
		UserID:     userID,
		Code:       code,
		DurationMs: durationMs,
		Rows:       len(batches),
		Success:    err == nil,
	}
	if err != nil {
		history.ErrorMessage = err.Error()
	}

	if historyErr := s.historyRepo.Insert(ctx, history); historyErr != nil {
		s.logger.Error("failed to store search history", "error", historyErr)
	}

	if err != nil {
		if errors.Is(err, domain.ErrWarehouseUnavailable) {
			s.logger.Error(
				"warehouse unavailable",
				"code", code,
				"normalized", parsed,
				"mode", mode,
				"duration_ms", durationMs,
				"error", err,
			)
		} else {
			s.logger.Error(
				"warehouse search failed",
				"code", code,
				"normalized", parsed,
				"mode", mode,
				"duration_ms", durationMs,
				"error", err,
			)
		}
	} else {
		s.logger.Info(
			"warehouse search done",
			"code", code,
			"normalized", parsed,
			"mode", mode,
			"duration_ms", durationMs,
			"rows", len(batches),
		)
	}

	return batches, err
}

var (
	reWithYear = regexp.MustCompile(`^\d{2}%\d{4,10}$`)
	reDigits   = regexp.MustCompile(`^\d{4,10}$`)
)

// parseCode implements DAX routing rules from spec.
func parseCode(code string) (string, domain.QueryMode, error) {
	raw := strings.TrimSpace(code)
	if raw == "" {
		return "", domain.QueryModeBase, ErrInvalidCode
	}

	mode := domain.QueryModeBase
	// Mode suffix has priority: * = color, @ = with user.
	switch {
	case strings.HasSuffix(raw, "*"):
		mode = domain.QueryModeWithColor
		raw = strings.TrimSuffix(raw, "*")
	case strings.HasSuffix(raw, "@"):
		mode = domain.QueryModeWithUser
		raw = strings.TrimSuffix(raw, "@")
	}

	// If input contains non-digits (letters/dashes) and no year format,
	// treat it as location (partial search).
	if strings.ContainsAny(raw, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz") ||
		strings.Contains(raw, "-") ||
		strings.Contains(raw, "_") {
		if mode == domain.QueryModeWithUser {
			return raw, domain.QueryModePartialWithUser, nil
		}
		return raw, domain.QueryModePartial, nil
	}

	if strings.Contains(raw, "%") {
		// Format with explicit year, e.g. 21%12345.
		if !reWithYear.MatchString(raw) {
			return "", mode, ErrInvalidCode
		}
		return raw, mode, nil
	}

	// Format without year (current year is injected).
	if !reDigits.MatchString(raw) {
		return "", mode, ErrInvalidCode
	}

	year := time.Now().Year() % 100
	return fmt.Sprintf("%02d%%%s", year, raw), mode, nil
}
