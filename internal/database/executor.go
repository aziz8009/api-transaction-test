package database

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DBExecutor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

// Implementasi DBExecutor untuk sqlx.DB
type SqlxDBExecutor struct {
	DB *sqlx.DB
}

func (e *SqlxDBExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return e.DB.ExecContext(ctx, query, args...)
}

func (e *SqlxDBExecutor) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return e.DB.QueryxContext(ctx, query, args...)
}

func (e *SqlxDBExecutor) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return e.DB.QueryRowxContext(ctx, query, args...)
}

func (e *SqlxDBExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.DB.GetContext(ctx, dest, query, args...)
}

func (e *SqlxDBExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.DB.SelectContext(ctx, dest, query, args...)
}

// Implementasi DBExecutor untuk sqlx.Tx
type SqlxTxExecutor struct {
	Tx *sqlx.Tx
}

func (e *SqlxTxExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return e.Tx.ExecContext(ctx, query, args...)
}

func (e *SqlxTxExecutor) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return e.Tx.QueryxContext(ctx, query, args...)
}

func (e *SqlxTxExecutor) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return e.Tx.QueryRowxContext(ctx, query, args...)
}

func (e *SqlxTxExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.Tx.GetContext(ctx, dest, query, args...)
}

func (e *SqlxTxExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return e.Tx.SelectContext(ctx, dest, query, args...)
}
