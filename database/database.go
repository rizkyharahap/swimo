package database

import (
	"context"
	"log/slog"
	"time"

	"github.com/rizkyharahap/swimo/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, cfg *config.Config) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		slog.Error("db parse config failed", slog.String("err", err.Error()))
		return nil, err
	}

	// defaults
	poolConfig.MaxConns = cfg.Database.MaxConns
	poolConfig.MinConns = cfg.Database.MinConns
	poolConfig.MaxConnLifetime = cfg.Database.MaxConnLifetime
	poolConfig.MaxConnIdleTime = cfg.Database.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		slog.Error("db pool create failed", slog.String("err", err.Error()))
		return nil, err
	}

	pgCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	err = pool.Ping(pgCtx)
	if err != nil {
		slog.Error("db ping failed", slog.String("err", err.Error()))
		pool.Close()
		return nil, err
	}

	slog.Info("db connected")
	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	if db.Pool != nil {
		slog.Info("db closing")
		db.Pool.Close()
	}
}
