package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetEnv_WhenEnvExists(t *testing.T) {
	t.Setenv("TEST_KEY", "test-value")

	value := getEnv("TEST_KEY", "default-value")

	assert.Equal(t, "test-value", value)

}

func TestGetEnv_WhenEnvNotExists(t *testing.T) {
	os.Unsetenv("TEST_KEY")

	value := getEnv("TEST_KEY", "default-value")

	assert.Equal(t, "default-value", value)

}

func TestGetEnv_WhenEnvEmpty(t *testing.T) {
	t.Setenv("TEST_KEY", "")

	value := getEnv("TEST_KEY", "default-value")

	assert.Equal(t, "default-value", value)

}

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/testdb")

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, "9000", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost:5432/testdb", cfg.DatabaseURL)

}

func TestLoad_WithDefaultValues(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	cfg := Load()

	assert.NotNil(t, cfg)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(
		t,
		"postgres://postgres:postgres@localhost:5432/sawitpro?sslmode=disable",
		cfg.DatabaseURL,
	)

}
