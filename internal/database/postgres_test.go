package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPostgresConnection_Success(t *testing.T) {
	dsn := "host=192.168.208.1 port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"

	db, err := NewPostgresConnection(dsn)

	assert.NoError(t, err)
	assert.NotNil(t, db)

	if db != nil {
		db.Close()
	}
}

func TestNewPostgresConnection_InvalidDSN(t *testing.T) {
	dsn := "invalid-dsn"

	db, err := NewPostgresConnection(dsn)

	assert.Error(t, err)
	assert.Nil(t, db)
}
