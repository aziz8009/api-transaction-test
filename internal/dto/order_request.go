package dto

import "time"

type GetOrdersRequest struct {
	Status    string     `query:"status"`
	StartDate *time.Time `query:"start_date"`
	EndDate   *time.Time `query:"end_date"`
	Page      int        `query:"page"`
	Limit     int        `query:"limit"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status" validate:"required"`
	Note   string `json:"note"`
}
