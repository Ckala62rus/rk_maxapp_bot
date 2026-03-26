package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// UserRepository is a Postgres implementation of domain.UserRepository.
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetByMaxUserID returns user by MAX user id.
func (r *UserRepository) GetByMaxUserID(ctx context.Context, maxUserID int64) (*domain.User, error) {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.GetByMaxUserID")
	span.SetAttributes(attribute.Int64("max_user_id", maxUserID))
	defer span.End()

	const query = `
SELECT id, max_user_id, max_username, max_first_name, max_last_name, language_code, photo_url,
       first_name, last_name, is_admin, is_blocked, is_approved, created_at, updated_at
FROM users
WHERE max_user_id = $1
`
	var user domain.User
	// Single-row query; return ErrNotFound for no rows.
	err := r.db.QueryRowContext(ctx, query, maxUserID).Scan(
		&user.ID,
		&user.MaxUserID,
		&user.MaxUsername,
		&user.MaxFirstName,
		&user.MaxLastName,
		&user.LanguageCode,
		&user.PhotoURL,
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin,
		&user.IsBlocked,
		&user.IsApproved,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// GetByID returns user by internal id.
func (r *UserRepository) GetByID(ctx context.Context, id int64) (*domain.User, error) {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.GetByID")
	span.SetAttributes(attribute.Int64("user_id", id))
	defer span.End()

	const query = `
SELECT id, max_user_id, max_username, max_first_name, max_last_name, language_code, photo_url,
       first_name, last_name, is_admin, is_blocked, is_approved, created_at, updated_at
FROM users
WHERE id = $1
`
	var user domain.User
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.MaxUserID,
		&user.MaxUsername,
		&user.MaxFirstName,
		&user.MaxLastName,
		&user.LanguageCode,
		&user.PhotoURL,
		&user.FirstName,
		&user.LastName,
		&user.IsAdmin,
		&user.IsBlocked,
		&user.IsApproved,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &user, nil
}

// Create inserts a new user and returns populated record.
func (r *UserRepository) Create(ctx context.Context, user domain.User) (*domain.User, error) {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.Create")
	span.SetAttributes(attribute.Int64("max_user_id", user.MaxUserID))
	defer span.End()

	const query = `
INSERT INTO users (max_user_id, max_username, max_first_name, max_last_name, language_code, photo_url,
                   first_name, last_name, is_admin, is_blocked, is_approved)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING id, created_at, updated_at
`
	// Use RETURNING to avoid extra SELECT.
	err := r.db.QueryRowContext(
		ctx,
		query,
		user.MaxUserID,
		user.MaxUsername,
		user.MaxFirstName,
		user.MaxLastName,
		user.LanguageCode,
		user.PhotoURL,
		user.FirstName,
		user.LastName,
		user.IsAdmin,
		user.IsBlocked,
		user.IsApproved,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateProfile updates first/last name for user.
func (r *UserRepository) UpdateProfile(ctx context.Context, id int64, firstName, lastName string) error {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.UpdateProfile")
	span.SetAttributes(attribute.Int64("user_id", id))
	defer span.End()

	const query = `
UPDATE users
SET first_name = $1, last_name = $2, updated_at = $3
WHERE id = $4
`
	_, err := r.db.ExecContext(ctx, query, firstName, lastName, time.Now().UTC(), id)
	return err
}

// UpdateFlags updates approval/admin/block flags with partial updates.
func (r *UserRepository) UpdateFlags(ctx context.Context, id int64, isApproved, isBlocked, isAdmin *bool) error {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.UpdateFlags")
	span.SetAttributes(attribute.Int64("user_id", id))
	defer span.End()

	const query = `
UPDATE users
SET is_approved = COALESCE($1, is_approved),
    is_blocked  = COALESCE($2, is_blocked),
    is_admin    = COALESCE($3, is_admin),
    updated_at  = $4
WHERE id = $5
`
	_, err := r.db.ExecContext(ctx, query, isApproved, isBlocked, isAdmin, time.Now().UTC(), id)
	return err
}

// List returns users sorted by creation time.
func (r *UserRepository) List(ctx context.Context) ([]domain.User, error) {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.List")
	defer span.End()

	const query = `
SELECT id, max_user_id, max_username, max_first_name, max_last_name, language_code, photo_url,
       first_name, last_name, is_admin, is_blocked, is_approved, created_at, updated_at
FROM users
ORDER BY created_at DESC
`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	// Iterate through all rows and build slice.
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.MaxUserID,
			&user.MaxUsername,
			&user.MaxFirstName,
			&user.MaxLastName,
			&user.LanguageCode,
			&user.PhotoURL,
			&user.FirstName,
			&user.LastName,
			&user.IsAdmin,
			&user.IsBlocked,
			&user.IsApproved,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

// SearchByName finds users by profile first/last name (case-insensitive).
func (r *UserRepository) SearchByName(ctx context.Context, query string) ([]domain.User, error) {
	tracer := otel.Tracer("repo.postgres")
	ctx, span := tracer.Start(ctx, "User.SearchByName")
	defer span.End()

	const querySQL = `
SELECT id, max_user_id, max_username, max_first_name, max_last_name, language_code, photo_url,
       first_name, last_name, is_admin, is_blocked, is_approved, created_at, updated_at
FROM users
WHERE first_name ILIKE $1
   OR last_name ILIKE $1
   OR (first_name || ' ' || last_name) ILIKE $1
ORDER BY created_at DESC
`
	pattern := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, querySQL, pattern)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.MaxUserID,
			&user.MaxUsername,
			&user.MaxFirstName,
			&user.MaxLastName,
			&user.LanguageCode,
			&user.PhotoURL,
			&user.FirstName,
			&user.LastName,
			&user.IsAdmin,
			&user.IsBlocked,
			&user.IsApproved,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}
