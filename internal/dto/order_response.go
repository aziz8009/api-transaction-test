package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type OrderListResponse struct {
	ID          uuid.UUID `json:"id"`
	OrderNumber string    `json:"order_number"`
	UserID      uuid.UUID `json:"user_id"`
	GrandTotal  float64   `json:"grand_total"`
	Status      string    `json:"status"`
	ItemsCount  int       `json:"items_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type OrderDetailResponse struct {
	ID              uuid.UUID           `json:"id"`
	OrderNumber     string              `json:"order_number"`
	UserID          uuid.UUID           `json:"user_id"`
	Items           []OrderItemResponse `json:"items"`
	Summary         CheckoutSummary     `json:"summary"`
	Payment         PaymentResponse     `json:"payment"`
	ShippingAddress string              `json:"shipping_address"`
	Status          string              `json:"status"`
	StatusHistory   []StatusHistory     `json:"status_history"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type StatusHistory struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func ToOrderListResponse(orders []model.Order, total, page, limit int) map[string]interface{} {
	var responses []OrderListResponse
	for _, o := range orders {
		responses = append(responses, OrderListResponse{
			ID:          o.ID,
			OrderNumber: o.OrderNumber,
			UserID:      o.UserID,
			GrandTotal:  o.GrandTotal,
			Status:      o.Status,
			CreatedAt:   o.CreatedAt,
		})
	}

	totalPages := (total + limit - 1) / limit

	return map[string]interface{}{
		"orders": responses,
		"pagination": PaginationResponse{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: limit,
		},
	}
}
