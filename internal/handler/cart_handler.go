package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/dto"
	"github.com/ifortepay/ApiBackendTest/internal/pkg/utils"
	"github.com/ifortepay/ApiBackendTest/internal/service"
	"github.com/labstack/echo/v4"
)

type CartHandler struct {
	cartService service.CartService
}

func NewCartHandler(cartService service.CartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// GetCart handles GET /api/cart
func (h *CartHandler) GetCart(c echo.Context) error {
	cart, err := h.cartService.GetActiveCart(c.Request().Context())
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get cart")
	}

	totalAmount, discountAmount, finalAmount, err := h.cartService.CalculateCartTotal(c.Request().Context(), cart)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate cart total")
	}

	summary := dto.CartSummaryResponse{
		Subtotal:      totalAmount,
		TotalDiscount: discountAmount,
		GrandTotal:    finalAmount,
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToCartResponse(cart, summary))
}

// AddToCart handles POST /api/cart
func (h *CartHandler) AddToCart(c echo.Context) error {
	var req dto.AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	cart, err := h.cartService.AddToCart(c.Request().Context(), req.ProductSKU, req.Quantity)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	totalAmount, discountAmount, finalAmount, err := h.cartService.CalculateCartTotal(c.Request().Context(), cart)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate cart total")
	}

	summary := dto.CartSummaryResponse{
		Subtotal:      totalAmount,
		TotalDiscount: discountAmount,
		GrandTotal:    finalAmount,
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToCartResponse(cart, summary))
}

// UpdateCartItem handles PUT /api/cart/:itemId
func (h *CartHandler) UpdateCartItem(c echo.Context) error {
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID")
	}

	var req dto.UpdateCartItemRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	cart, err := h.cartService.UpdateCartItem(c.Request().Context(), itemID, req.Quantity)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	totalAmount, discountAmount, finalAmount, err := h.cartService.CalculateCartTotal(c.Request().Context(), cart)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate cart total")
	}

	summary := dto.CartSummaryResponse{
		Subtotal:      totalAmount,
		TotalDiscount: discountAmount,
		GrandTotal:    finalAmount,
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToCartResponse(cart, summary))
}

// RemoveCartItem handles DELETE /api/cart/:itemId
func (h *CartHandler) RemoveCartItem(c echo.Context) error {
	itemID, err := uuid.Parse(c.Param("itemId"))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid item ID")
	}

	cart, err := h.cartService.RemoveCartItem(c.Request().Context(), itemID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	totalAmount, discountAmount, finalAmount, err := h.cartService.CalculateCartTotal(c.Request().Context(), cart)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to calculate cart total")
	}

	summary := dto.CartSummaryResponse{
		Subtotal:      totalAmount,
		TotalDiscount: discountAmount,
		GrandTotal:    finalAmount,
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToCartResponse(cart, summary))
}
