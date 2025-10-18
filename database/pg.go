package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mwdev22/core/config"
)

func NewPostgres(cfg *config.DatabaseConfig) (*pgxpool.Pool, error) {
	configStr := cfg.URI

	conf, err := pgxpool.ParseConfig(configStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse db config: %w", err)
	}

	conf.MaxConns = int32(cfg.MaxOpenConns)
	conf.MinConns = int32(cfg.MinIdleConns)
	conf.MaxConnLifetime = time.Duration(cfg.ConnMaxLifetime) * time.Minute

	dbpool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create db pool: %w", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping db: %w", err)
	}

	return dbpool, nil
}
