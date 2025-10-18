package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // postgres driver
	"github.com/mwdev22/core/config"
)

func NewSqlx(cfg *config.DatabaseConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlx db: %w", err)
	}

	// configure connection pool
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Minute)

	// test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping sqlx db: %w", err)
	}

	return db, nil
}
