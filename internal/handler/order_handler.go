package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/dto"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/pkg/utils"
	"github.com/ifortepay/ApiBackendTest/internal/service"
	"github.com/labstack/echo/v4"
)

type OrderHandler struct {
	checkoutService service.CheckoutService
}

func NewOrderHandler(checkoutService service.CheckoutService) *OrderHandler {
	return &OrderHandler{
		checkoutService: checkoutService,
	}
}

// GetOrders handles GET /api/orders
func (h *OrderHandler) GetOrders(c echo.Context) error {
	var req dto.GetOrdersRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
	}

	filter := model.OrderFilter{
		Status:    req.Status,
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
		Page:      req.Page,
		Limit:     req.Limit,
	}

	orders, total, err := h.checkoutService.GetOrders(c.Request().Context(), filter)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get orders")
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToOrderListResponse(orders, total, filter.Page, filter.Limit))
}

// GetOrderByID handles GET /api/orders/:id
func (h *OrderHandler) GetOrderByID(c echo.Context) error {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	order, err := h.checkoutService.GetOrder(c.Request().Context(), orderID)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Order not found")
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
		Status:        "success",
		Amount:        order.GrandTotal,
		PaymentMethod: "credit_card",
		TransactionID: "txn_" + uuid.New().String()[:8],
	}

	orderResp := dto.OrderDetailResponse{
		ID:              order.ID,
		OrderNumber:     order.OrderNumber,
		UserID:          order.UserID,
		Items:           items,
		Summary:         summary,
		Payment:         payment,
		ShippingAddress: order.ShippingAddress,
		Status:          order.Status,
		StatusHistory: []dto.StatusHistory{
			{Status: "pending_payment", Timestamp: order.CreatedAt},
			{Status: order.Status, Timestamp: order.UpdatedAt},
		},
		CreatedAt: order.CreatedAt,
		UpdatedAt: order.UpdatedAt,
	}

	return utils.SuccessResponse(c, http.StatusOK, orderResp)
}

// UpdateOrderStatus handles PUT /api/orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c echo.Context) error {
	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid order ID")
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request body")
	}

	if err := utils.ValidateStruct(req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	if err := h.checkoutService.UpdateOrderStatus(c.Request().Context(), orderID, req.Status); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	return utils.SuccessResponse(c, http.StatusOK, map[string]interface{}{
		"order_id": orderID.String(),
		"status":   req.Status,
		"message":  "Order status updated successfully",
	})
}
