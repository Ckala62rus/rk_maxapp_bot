// Package postgres provides PostgreSQL connection and repositories.
package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"maxapp/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// New opens a PostgreSQL connection using pgx stdlib driver.
func New(cfg config.PostgresConfig) (*sql.DB, error) {
	// Build DSN from config to avoid leaking credentials in logs.
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode,
	)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	// Reasonable defaults for connection pool.
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)

	// Ping to verify connectivity during startup.
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
