package db

import (
	"context"
	"database/sql"
)

type txKey struct{}

type postgresUnitOfWork struct {
	db *sql.DB
}

func NewUnitOfWork(db *sql.DB) *postgresUnitOfWork {
	return &postgresUnitOfWork{db: db}
}

func (t *postgresUnitOfWork) Do(ctx context.Context, fn func(ctx context.Context) error) error {
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
