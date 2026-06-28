package database

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	cleanup := func() {
		db.Close()
	}

	return sqlxDB, mock, cleanup
}

func TestSqlxDBExecutor_ExecContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	executor := &SqlxDBExecutor{DB: db}

	mock.ExpectExec("INSERT INTO products").
		WithArgs(1).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err := executor.ExecContext(
		context.Background(),
		"INSERT INTO products",
		1,
	)

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxDBExecutor_QueryxContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	executor := &SqlxDBExecutor{DB: db}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	result, err := executor.QueryxContext(
		context.Background(),
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxDBExecutor_QueryRowxContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	executor := &SqlxDBExecutor{DB: db}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(10)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var id int

	err := executor.
		QueryRowxContext(context.Background(), "SELECT id FROM products").
		Scan(&id)

	require.NoError(t, err)
	assert.Equal(t, 10, id)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxDBExecutor_GetContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	executor := &SqlxDBExecutor{DB: db}

	type Product struct {
		ID int `db:"id"`
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(100)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var p Product

	err := executor.GetContext(
		context.Background(),
		&p,
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	assert.Equal(t, 100, p.ID)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxDBExecutor_SelectContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	executor := &SqlxDBExecutor{DB: db}

	type Product struct {
		ID int `db:"id"`
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var products []Product

	err := executor.SelectContext(
		context.Background(),
		&products,
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	assert.Len(t, products, 2)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxTxExecutor_ExecContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()

	tx, err := db.Beginx()
	require.NoError(t, err)

	executor := &SqlxTxExecutor{
		Tx: tx,
	}

	mock.ExpectExec("UPDATE products").
		WithArgs(100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	_, err = executor.ExecContext(
		context.Background(),
		"UPDATE products",
		100,
	)

	require.NoError(t, err)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxTxExecutor_QueryxContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()

	tx, err := db.Beginx()
	require.NoError(t, err)

	executor := &SqlxTxExecutor{
		Tx: tx,
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	result, err := executor.QueryxContext(
		context.Background(),
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	require.NotNil(t, result)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxTxExecutor_QueryRowxContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()

	tx, err := db.Beginx()
	require.NoError(t, err)

	executor := &SqlxTxExecutor{
		Tx: tx,
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(5)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var id int

	err = executor.
		QueryRowxContext(context.Background(), "SELECT id FROM products").
		Scan(&id)

	require.NoError(t, err)
	assert.Equal(t, 5, id)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxTxExecutor_GetContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()

	tx, err := db.Beginx()
	require.NoError(t, err)

	executor := &SqlxTxExecutor{
		Tx: tx,
	}

	type Product struct {
		ID int `db:"id"`
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(9)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var product Product

	err = executor.GetContext(
		context.Background(),
		&product,
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	assert.Equal(t, 9, product.ID)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSqlxTxExecutor_SelectContext(t *testing.T) {
	db, mock, cleanup := newMockDB(t)
	defer cleanup()

	mock.ExpectBegin()

	tx, err := db.Beginx()
	require.NoError(t, err)

	executor := &SqlxTxExecutor{
		Tx: tx,
	}

	type Product struct {
		ID int `db:"id"`
	}

	rows := sqlmock.NewRows([]string{"id"}).
		AddRow(1).
		AddRow(2)

	mock.ExpectQuery("SELECT id FROM products").
		WillReturnRows(rows)

	var products []Product

	err = executor.SelectContext(
		context.Background(),
		&products,
		"SELECT id FROM products",
	)

	require.NoError(t, err)
	assert.Len(t, products, 2)

	mock.ExpectCommit()
	require.NoError(t, tx.Commit())

	assert.NoError(t, mock.ExpectationsWereMet())
}
