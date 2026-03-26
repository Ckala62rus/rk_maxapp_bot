package domain

import (
	"context"
	"errors"
)

// ErrNotFound is returned by repositories when record doesn't exist.
var ErrNotFound = errors.New("not found")

// ErrWarehouseUnavailable indicates DAX/MSSQL is not reachable.
var ErrWarehouseUnavailable = errors.New("warehouse unavailable")

// UserRepository defines CRUD operations for users.
type UserRepository interface {
	GetByMaxUserID(ctx context.Context, maxUserID int64) (*User, error)
	GetByID(ctx context.Context, id int64) (*User, error)
	Create(ctx context.Context, user User) (*User, error)
	UpdateProfile(ctx context.Context, id int64, firstName, lastName string) error
	UpdateFlags(ctx context.Context, id int64, isApproved, isBlocked, isAdmin *bool) error
	List(ctx context.Context) ([]User, error)
	SearchByName(ctx context.Context, query string) ([]User, error)
}

// SearchHistoryRepository stores audit search records.
type SearchHistoryRepository interface {
	Insert(ctx context.Context, history SearchHistory) error
}

// WarehouseRepository queries DAX batches.
type WarehouseRepository interface {
	FindBatches(ctx context.Context, number string, mode QueryMode) ([]Batch, error)
}

// QueryMode selects SQL variant for DAX lookup.
type QueryMode int

// Supported DAX query modes.
const (
	QueryModeBase QueryMode = iota
	QueryModeWithColor
	QueryModeWithUser
	QueryModePartial
	QueryModePartialWithUser
)
