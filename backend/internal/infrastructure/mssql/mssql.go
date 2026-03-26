// Package mssql provides MSSQL connection and repositories.
package mssql

import (
	"database/sql"
	"fmt"
	"net/url"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/config"

	_ "github.com/denisenkom/go-mssqldb"
)

// New opens MSSQL connection using go-mssqldb driver.
func New(cfg config.MSSQLConfig) (*sql.DB, error) {
	// Build DSN with query parameters to avoid plaintext concatenation.
	query := url.Values{}
	query.Add("database", cfg.Database)
	query.Add("encrypt", cfg.Encrypt)

	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%s?%s",
		url.QueryEscape(cfg.User),
		url.QueryEscape(cfg.Password),
		cfg.Host,
		cfg.Port,
		query.Encode(),
	)

	db, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, err
	}

	// Connection pool settings.
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Ping to ensure startup fails fast on bad config.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
