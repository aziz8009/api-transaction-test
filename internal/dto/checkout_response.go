package dto

import (
	"time"

	"github.com/google/uuid"
)

type OrderItemResponse struct {
	SKU              string  `json:"sku"`
	Name             string  `json:"name"`
	Quantity         int     `json:"quantity"`
	UnitPrice        float64 `json:"unit_price"`
	OriginalSubtotal float64 `json:"original_subtotal"`
	Discount         float64 `json:"discount"`
	FinalPrice       float64 `json:"final_price"`
	PromotionApplied string  `json:"promotion_applied,omitempty"`
}

type PromotionBreakdown struct {
	Promotion     string   `json:"promotion"`
	Saving        float64  `json:"saving"`
	ItemsAffected []string `json:"items_affected"`
}

type CheckoutSummary struct {
	Subtotal           float64              `json:"subtotal"`
	TotalDiscount      float64              `json:"total_discount"`
	GrandTotal         float64              `json:"grand_total"`
	PromotionBreakdown []PromotionBreakdown `json:"promotion_breakdown"`
}

type PaymentResponse struct {
	PaymentID     string  `json:"payment_id"`
	Status        string  `json:"status"`
	Amount        float64 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	TransactionID string  `json:"transaction_id"`
}

type CheckoutResponse struct {
	OrderID         uuid.UUID           `json:"order_id"`
	OrderNumber     string              `json:"order_number"`
	Items           []OrderItemResponse `json:"items"`
	Summary         CheckoutSummary     `json:"summary"`
	Payment         PaymentResponse     `json:"payment"`
	ShippingAddress string              `json:"shipping_address"`
	Status          string              `json:"status"`
	CreatedAt       time.Time           `json:"created_at"`
}
