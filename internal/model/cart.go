package model

import (
	"time"

	"github.com/google/uuid"
)

type Cart struct {
	ID             uuid.UUID `db:"id" json:"id"`
	TotalAmount    float64   `db:"total_amount" json:"total_amount"`
	DiscountAmount float64   `db:"discount_amount" json:"discount_amount"`
	FinalAmount    float64   `db:"final_amount" json:"final_amount"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type CartItem struct {
	ID             uuid.UUID `db:"id" json:"id"`
	CartID         uuid.UUID `db:"cart_id" json:"cart_id"`
	ProductSKU     string    `db:"product_sku" json:"product_sku"`
	Quantity       int       `db:"quantity" json:"quantity"`
	UnitPrice      float64   `db:"unit_price" json:"unit_price"`
	TotalPrice     float64   `db:"total_price" json:"total_price"`
	DiscountAmount float64   `db:"discount_amount" json:"discount_amount"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type CartWithItems struct {
	Cart
	Items []CartItemDetail `json:"items"`
}

type CartItemDetail struct {
	ID               uuid.UUID `json:"id"`
	CartID           uuid.UUID `json:"cart_id"`
	ProductSKU       string    `json:"product_sku"`
	ProductName      string    `json:"product_name"`
	Quantity         int       `json:"quantity"`
	UnitPrice        float64   `json:"unit_price"`
	TotalPrice       float64   `json:"total_price"`
	DiscountAmount   float64   `json:"discount_amount"`
	FinalPrice       float64   `json:"final_price"`
	PromotionApplied string    `json:"promotion_applied,omitempty"`
}
