package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/dto"
	"github.com/ifortepay/ApiBackendTest/internal/pkg/utils"
	"github.com/ifortepay/ApiBackendTest/internal/service"
	"github.com/labstack/echo/v4"
)

type CheckoutHandler struct {
	checkoutService service.CheckoutService
}

func NewCheckoutHandler(checkoutService service.CheckoutService) *CheckoutHandler {
	return &CheckoutHandler{
		checkoutService: checkoutService,
	}
}

// Checkout handles POST /api/checkout
func (h *CheckoutHandler) Checkout(c echo.Context) error {
	var req dto.CheckoutRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	cartID, err := uuid.Parse(req.CartID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid cart ID")
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid user ID")
	}

	order, err := h.checkoutService.ProcessCheckout(
		c.Request().Context(),
		cartID,
		userID,
		req.ShippingAddress,
		req.PaymentMethod,
		req.IdempotencyKey,
	)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	// Build response
	var items []dto.OrderItemResponse
	for _, item := range order.Items {
		items = append(items, dto.OrderItemResponse{
			SKU:              item.ProductSKU,
			Quantity:         item.Quantity,
			UnitPrice:        item.UnitPrice,
			OriginalSubtotal: item.UnitPrice * float64(item.Quantity),
			Discount:         item.Discount,
			FinalPrice:       item.FinalPrice,
			PromotionApplied: item.PromotionApplied,
		})
	}

	summary := dto.CheckoutSummary{
		Subtotal:      order.GrandTotal + order.DiscountTotal,
		TotalDiscount: order.DiscountTotal,
		GrandTotal:    order.GrandTotal,
	}

	// Mock payment response
	payment := dto.PaymentResponse{
		PaymentID:     uuid.New().String(),
		Status:        "pending",
		Amount:        order.GrandTotal,
		PaymentMethod: "credit_card",
		TransactionID: "txn_" + uuid.New().String()[:8],
	}

	checkoutResp := dto.CheckoutResponse{
		OrderID:         order.ID,
		OrderNumber:     order.OrderNumber,
		Items:           items,
		Summary:         summary,
		Payment:         payment,
		ShippingAddress: order.ShippingAddress,
		Status:          order.Status,
		CreatedAt:       order.CreatedAt,
	}

	return utils.SuccessResponse(c, http.StatusOK, checkoutResp)
}

func (h *CheckoutHandler) ConfirmPayment(c echo.Context) error {
	var req dto.PaymentConfirmRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	orderID, err := uuid.Parse(req.OrderID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	if err := h.checkoutService.ConfirmPayment(c.Request().Context(), orderID, req.PaymentStatus, req.TransactionReference); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"order_id": req.OrderID,
		"status":   "success",
		"message":  "Payment confirmed successfully",
	})
}
