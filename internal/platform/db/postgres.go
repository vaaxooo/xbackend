package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Postgres struct {
	DB *sql.DB
}

func OpenPostgres(dsn string, maxOpen, maxIdle int, connMaxLifetime time.Duration) (*sql.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres dsn is empty")
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("sql.Open: %w", err)
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(connMaxLifetime)

	// fail-fast: сразу проверяем соединение
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return db, nil
}
