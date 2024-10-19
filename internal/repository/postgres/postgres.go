package postgres

import (
	"context"
	"fmt"
	"kuchak/internal/config"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresSession() (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.AppConfig.PostgresUser,
		config.AppConfig.PostgresPasswrod,
		config.AppConfig.PostgresHost,
		config.AppConfig.PostgresPort,
		config.AppConfig.PostgresDB)

	conf, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres config: %w", err)
	}

	conf.MaxConns = 10
	conf.MaxConnLifetime = time.Hour
	conf.MinConns = 2

	pool, err := pgxpool.NewWithConfig(context.Background(), conf)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return pool, nil
}
