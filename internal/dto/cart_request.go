package dto

type AddToCartRequest struct {
	ProductSKU string `json:"product_sku" validate:"required"`
	Quantity   int    `json:"quantity" validate:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}
