package dto

type GetProductsRequest struct {
	Search   string  `query:"search"`
	MinPrice float64 `query:"min_price"`
	MaxPrice float64 `query:"max_price"`
	SKU      string  `query:"sku"`
	Page     int     `query:"page"`
	Limit    int     `query:"limit"`
}
