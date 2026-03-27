// Package postgres provides PostgreSQL connection and repositories.
package postgres

import (
	"database/sql"
	"net"
	"net/url"
	"time"

	"github.com/Ckala62rus/rk_maxapp_bot/backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// New opens a PostgreSQL connection using pgx stdlib driver.
func New(cfg config.PostgresConfig) (*sql.DB, error) {
	// Build URI with net/url so user/password with @, :, %, etc. are encoded (pgx parses a URL).
	u := &url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   net.JoinHostPort(cfg.Host, cfg.Port),
		Path:   "/" + url.PathEscape(cfg.Database),
	}
	q := url.Values{}
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()
	dsn := u.String()
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
