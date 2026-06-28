package routes

import (
	"net/http"

	"github.com/ifortepay/ApiBackendTest/internal/handler"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	CheckoutHandler *handler.CheckoutHandler
	ProductHandler  *handler.ProductHandler
	CartHandler     *handler.CartHandler
	OrderHandler    *handler.OrderHandler
}

func Register(e *echo.Echo, h *Handler) {
	e.GET("/health", healthCheck)

	v1 := e.Group("/api/v1")

	registerProductRoutes(v1, h)
	registerCartRoutes(v1, h)
	registerCheckOutRoutes(v1, h)
	registerPaymentRoutes(v1, h)
	registerOrderRoutes(v1, h)
}

func registerProductRoutes(g *echo.Group, h *Handler) {
	productGroup := g.Group("/products")

	productGroup.GET("", h.ProductHandler.GetProducts)
	productGroup.GET("/:sku", h.ProductHandler.GetProductBySKU)
}

func registerCartRoutes(g *echo.Group, h *Handler) {
	cartGroup := g.Group("/carts")
	cartGroup.GET("", h.CartHandler.GetCart)
	cartGroup.POST("/items", h.CartHandler.AddToCart)
	cartGroup.PUT("/items/:itemId", h.CartHandler.UpdateCartItem)
	cartGroup.DELETE("/items/:itemId", h.CartHandler.RemoveCartItem)
}

func registerCheckOutRoutes(g *echo.Group, h *Handler) {
	checkoutGroup := g.Group("/checkout")
	checkoutGroup.POST("", h.CheckoutHandler.Checkout)
}

func registerPaymentRoutes(g *echo.Group, h *Handler) {
	paymentGroup := g.Group("/payment")
	paymentGroup.POST("/confirm", h.CheckoutHandler.ConfirmPayment)
}

func registerOrderRoutes(g *echo.Group, h *Handler) {
	ordersGroup := g.Group("/orders")
	ordersGroup.GET("", h.OrderHandler.GetOrders)
	ordersGroup.GET("/:id", h.OrderHandler.GetOrderByID)
	ordersGroup.PUT("/:id/status", h.OrderHandler.UpdateOrderStatus)
}

func healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status": "ok",
	})
}
