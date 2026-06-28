package routes

import (
	"testing"

	"github.com/ifortepay/ApiBackendTest/internal/handler"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	e := echo.New()

	h := &Handler{
		ProductHandler:  &handler.ProductHandler{},
		CartHandler:     &handler.CartHandler{},
		CheckoutHandler: &handler.CheckoutHandler{},
		OrderHandler:    &handler.OrderHandler{},
	}

	Register(e, h)

	expectedRoutes := []struct {
		Method string
		Path   string
	}{
		{"GET", "/health"},

		{"GET", "/api/v1/products"},
		{"GET", "/api/v1/products/:sku"},

		{"GET", "/api/v1/carts"},
		{"POST", "/api/v1/carts/items"},
		{"PUT", "/api/v1/carts/items/:itemId"},
		{"DELETE", "/api/v1/carts/items/:itemId"},

		{"POST", "/api/v1/checkout"},

		{"POST", "/api/v1/payment/confirm"},

		{"GET", "/api/v1/orders"},
		{"GET", "/api/v1/orders/:id"},
		{"PUT", "/api/v1/orders/:id/status"},
	}

	routes := e.Routes()

	for _, expected := range expectedRoutes {
		found := false

		for _, route := range routes {
			if route.Method == expected.Method && route.Path == expected.Path {
				found = true
				break
			}
		}

		assert.Truef(
			t,
			found,
			"route %s %s is not registered",
			expected.Method,
			expected.Path,
		)
	}
}
