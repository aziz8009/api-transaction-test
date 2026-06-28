package model

import (
	"time"
)

type Product struct {
	SKU           string    `db:"sku" json:"sku"`
	Name          string    `db:"name" json:"name"`
	Price         float64   `db:"price" json:"price"`
	StockQuantity int       `db:"stock_quantity" json:"stock_quantity"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type ProductFilter struct {
	Search   string  `query:"status"`
	MinPrice float64 `query:"min_price"`
	MaxPrice float64 `query:"max_price"`
	SKU      string  `query:"sku"`
	Page     int     `query:"page"`
	Limit    int     `query:"limit"`
}
