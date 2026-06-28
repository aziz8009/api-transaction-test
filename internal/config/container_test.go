package config

import (
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ifortepay/ApiBackendTest/internal/database"
)

func TestBuildContainer_Success(t *testing.T) {
	original := newPostgresConnection
	defer func() {
		newPostgresConnection = original
	}()

	newPostgresConnection = func(dsn string) (*database.PostgresDB, error) {
		db := sqlx.NewDb(nil, "postgres")

		return &database.PostgresDB{
			DB: db,
		}, nil
	}

	cfg := &Config{
		DatabaseURL: "postgres://test",
	}

	container, err := BuildContainer(cfg)

	require.NoError(t, err)
	require.NotNil(t, container)

	assert.NotNil(t, container.ProductHandler)
	assert.NotNil(t, container.CartHandler)
	assert.NotNil(t, container.CheckoutHandler)
	assert.NotNil(t, container.OrderHandler)
}

func TestBuildContainer_DBConnectionError(t *testing.T) {
	original := newPostgresConnection
	defer func() {
		newPostgresConnection = original
	}()

	expectedErr := errors.New("database connection failed")

	newPostgresConnection = func(dsn string) (*database.PostgresDB, error) {
		return nil, expectedErr
	}

	cfg := &Config{
		DatabaseURL: "postgres://test",
	}

	container, err := BuildContainer(cfg)

	require.Error(t, err)
	assert.Nil(t, container)
	assert.Equal(t, expectedErr, err)
}
