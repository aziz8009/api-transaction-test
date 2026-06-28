package database

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTransactionMockDB(t *testing.T) (*sqlx.DB, sqlmock.Sqlmock, func()) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	cleanup := func() {
		_ = db.Close()
	}

	return sqlxDB, mock, cleanup
}

func TestNewTransactionManager(t *testing.T) {
	db, _, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	assert.NotNil(t, manager)
	assert.IsType(t, &transactionManager{}, manager)
}

func TestTransactionManager_Do_Success(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := manager.Do(context.Background(), func(tx *sqlx.Tx) error {
		return nil
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Do_Error(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	expectedErr := errors.New("callback error")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := manager.Do(context.Background(), func(tx *sqlx.Tx) error {
		return expectedErr
	})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Do_BeginError(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	expectedErr := errors.New("begin failed")

	mock.ExpectBegin().WillReturnError(expectedErr)

	err := manager.Do(context.Background(), func(tx *sqlx.Tx) error {
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_Do_Panic(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.Panics(t, func() {
		_ = manager.Do(context.Background(), func(tx *sqlx.Tx) error {
			panic("panic test")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_DoWithExecutor_Success(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	mock.ExpectBegin()
	mock.ExpectCommit()

	err := manager.DoWithExecutor(context.Background(), func(executor DBExecutor) error {
		assert.NotNil(t, executor)
		assert.IsType(t, &SqlxTxExecutor{}, executor)
		return nil
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_DoWithExecutor_Error(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	expectedErr := errors.New("executor error")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := manager.DoWithExecutor(context.Background(), func(executor DBExecutor) error {
		return expectedErr
	})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_DoWithExecutor_BeginError(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	expectedErr := errors.New("begin failed")

	mock.ExpectBegin().WillReturnError(expectedErr)

	err := manager.DoWithExecutor(context.Background(), func(executor DBExecutor) error {
		return nil
	})

	require.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionManager_DoWithExecutor_Panic(t *testing.T) {
	db, mock, cleanup := newTransactionMockDB(t)
	defer cleanup()

	manager := NewTransactionManager(db)

	mock.ExpectBegin()
	mock.ExpectRollback()

	assert.Panics(t, func() {
		_ = manager.DoWithExecutor(context.Background(), func(executor DBExecutor) error {
			panic("panic executor")
		})
	})

	assert.NoError(t, mock.ExpectationsWereMet())
}
