package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type TransactionManager interface {
	Do(ctx context.Context, fn func(tx *sqlx.Tx) error) error
	DoWithExecutor(ctx context.Context, fn func(executor DBExecutor) error) error
}

type transactionManager struct {
	db *sqlx.DB
}

func NewTransactionManager(db *sqlx.DB) TransactionManager {
	return &transactionManager{
		db: db,
	}
}

func (t *transactionManager) Do(
	ctx context.Context,
	fn func(tx *sqlx.Tx) error,
) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		fmt.Println("aziz1" + err.Error())
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	if err := fn(tx); err != nil {
		fmt.Println("aziz" + err.Error())
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (t *transactionManager) DoWithExecutor(
	ctx context.Context,
	fn func(executor DBExecutor) error,
) error {
	tx, err := t.db.BeginTxx(ctx, nil)
	if err != nil {
		fmt.Println("aziz2" + err.Error())
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			_ = tx.Rollback()
			panic(r)
		}
	}()

	executor := &SqlxTxExecutor{Tx: tx}
	if err := fn(executor); err != nil {
		fmt.Println("aziz3" + err.Error())
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
