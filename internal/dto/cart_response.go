package dto

import (
	"time"

	"github.com/google/uuid"
	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type CartItemResponse struct {
	ID               uuid.UUID `json:"id"`
	ProductSKU       string    `json:"product_sku"`
	ProductName      string    `json:"product_name"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	TotalPrice       float64   `json:"total_price"`
	DiscountAmount   float64   `json:"discount_amount"`
	FinalPrice       float64   `json:"final_price"`
	PromotionApplied string    `json:"promotion_applied,omitempty"`
}

type CartSummaryResponse struct {
	Subtotal      float64 `json:"subtotal"`
	TotalDiscount float64 `json:"total_discount"`
	GrandTotal    float64 `json:"grand_total"`
}

type CartResponse struct {
	CartID    uuid.UUID           `json:"cart_id"`
	Items     []CartItemResponse  `json:"items"`
	Summary   CartSummaryResponse `json:"summary"`
	Status    string              `json:"status"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type AddToCartResponse struct {
	CartID  uuid.UUID        `json:"cart_id"`
	Message string           `json:"message"`
	Item    CartItemResponse `json:"item"`
}

func ToCartItemResponse(item model.CartItemDetail) CartItemResponse {
	return CartItemResponse{
		ID:               item.ID,
		ProductSKU:       item.ProductSKU,
		ProductName:      item.ProductName,
		Quantity:         item.Quantity,
		UnitPrice:        item.UnitPrice,
		TotalPrice:       item.TotalPrice,
		DiscountAmount:   item.DiscountAmount,
		FinalPrice:       item.FinalPrice,
		PromotionApplied: item.PromotionApplied,
	}
}

func ToCartResponse(cart *model.CartWithItems, summary CartSummaryResponse) CartResponse {
	var items []CartItemResponse
	for _, item := range cart.Items {
		items = append(items, ToCartItemResponse(item))
	}

	return CartResponse{
		CartID:    cart.ID,
		Items:     items,
		Summary:   summary,
		Status:    cart.Status,
		CreatedAt: cart.CreatedAt,
		UpdatedAt: cart.UpdatedAt,
	}
}
