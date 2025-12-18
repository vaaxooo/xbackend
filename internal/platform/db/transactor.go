package db

import (
	"context"
	"database/sql"
)

type txKey struct{}

// Transactor provides a simple Unit of Work boundary.
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type postgresTransactor struct {
	db *sql.DB
}

func NewTransactor(db *sql.DB) Transactor {
	return &postgresTransactor{db: db}
}

func (t *postgresTransactor) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := t.db.BeginTx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}

	ctxWithTx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(ctxWithTx); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Executor returns *sql.Tx if present in ctx; otherwise returns *sql.DB.
func Executor(ctx context.Context, db *sql.DB) executor {
	if v := ctx.Value(txKey{}); v != nil {
		if tx, ok := v.(*sql.Tx); ok {
			return tx
		}
	}
	return db
}

// executor is the minimal interface shared by *sql.DB and *sql.Tx.
type executor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
}
