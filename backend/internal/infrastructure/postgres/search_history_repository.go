package postgres

import (
	"context"
	"database/sql"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// SearchHistoryRepository implements domain.SearchHistoryRepository.
type SearchHistoryRepository struct {
	db *sql.DB
}

// NewSearchHistoryRepository creates a new repository.
func NewSearchHistoryRepository(db *sql.DB) *SearchHistoryRepository {
	return &SearchHistoryRepository{db: db}
}

// Insert stores a search history record.
func (r *SearchHistoryRepository) Insert(ctx context.Context, history domain.SearchHistory) error {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "SearchHistory.Insert")
	span.SetAttributes(
		attribute.Int64("user_id", history.UserID),
		attribute.String("code", history.Code),
		attribute.Bool("success", history.Success),
	)
	defer span.End()

	const query = `
INSERT INTO search_history (user_id, code, duration_ms, rows, success, error_message)
VALUES ($1, $2, $3, $4, $5, $6)
`
	// Use nullable error_message to avoid empty string confusion.
	_, err := r.db.ExecContext(
		ctx,
		query,
		history.UserID,
		history.Code,
		history.DurationMs,
		history.Rows,
		history.Success,
		nullableString(history.ErrorMessage),
	)
	return err
}

// nullableString maps empty string to SQL NULL.
func nullableString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: value, Valid: true}
}
