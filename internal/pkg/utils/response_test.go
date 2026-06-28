package utils

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSuccessResponse(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	data := map[string]interface{}{
		"id":   1,
		"name": "Product A",
	}

	err := SuccessResponse(c, http.StatusOK, data)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var resp APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "success", resp.Status)
	assert.Empty(t, resp.Message)

	respData := resp.Data.(map[string]interface{})
	assert.Equal(t, float64(1), respData["id"])
	assert.Equal(t, "Product A", respData["name"])
}

func TestErrorResponse(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	c := e.NewContext(req, rec)

	err := ErrorResponse(c, http.StatusBadRequest, "invalid request")

	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var resp APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &resp)
	require.NoError(t, err)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "invalid request", resp.Message)
	assert.Nil(t, resp.Data)
}
