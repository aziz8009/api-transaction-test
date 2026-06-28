package dto

type CheckoutRequest struct {
	CartID          string `json:"cart_id" validate:"required"`
	UserID          string `json:"user_id" validate:"required"`
	PaymentMethod   string `json:"payment_method" validate:"required"`
	ShippingAddress string `json:"shipping_address" validate:"required"`
	IdempotencyKey  string `json:"idempotency_key" validate:"required"`
}

type PaymentConfirmRequest struct {
	OrderID              string `json:"order_id" validate:"required"`
	PaymentID            string `json:"payment_id" validate:"required"`
	PaymentStatus        string `json:"payment_status" validate:"required"`
	TransactionReference string `json:"transaction_reference" validate:"required"`
}
