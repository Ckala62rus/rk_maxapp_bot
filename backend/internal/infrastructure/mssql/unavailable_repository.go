package mssql

import (
	"context"
	"fmt"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/domain"
)

// unavailableWarehouseRepository returns a stable error when MSSQL is down.
type unavailableWarehouseRepository struct {
	cause error
}

// NewUnavailableWarehouseRepository builds a stub repository for dev mode.
func NewUnavailableWarehouseRepository(cause error) *unavailableWarehouseRepository {
	return &unavailableWarehouseRepository{cause: cause}
}

// FindBatches always returns an error when MSSQL is unavailable.
func (r *unavailableWarehouseRepository) FindBatches(ctx context.Context, number string, mode domain.QueryMode) ([]domain.Batch, error) {
	if r.cause != nil {
		return nil, fmt.Errorf("%w: %v", domain.ErrWarehouseUnavailable, r.cause)
	}
	return nil, domain.ErrWarehouseUnavailable
}
