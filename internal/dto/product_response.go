package dto

import (
	"time"

	"github.com/ifortepay/ApiBackendTest/internal/model"
)

type ProductResponse struct {
	SKU           string    `json:"sku"`
	Name          string    `json:"name"`
	Price         float64   `json:"price"`
	StockQuantity int       `json:"stock_quantity"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ProductListResponse struct {
	Products   []ProductResponse  `json:"products"`
	Pagination PaginationResponse `json:"pagination"`
}

type PaginationResponse struct {
	CurrentPage  int `json:"current_page"`
	TotalPages   int `json:"total_pages"`
	TotalItems   int `json:"total_items"`
	ItemsPerPage int `json:"items_per_page"`
}

func ToProductResponse(product *model.Product) ProductResponse {
	return ProductResponse{
		SKU:           product.SKU,
		Name:          product.Name,
		Price:         product.Price,
		StockQuantity: product.StockQuantity,
		CreatedAt:     product.CreatedAt,
		UpdatedAt:     product.UpdatedAt,
	}
}

func ToProductListResponse(products []model.Product, total, page, limit int) ProductListResponse {
	var responses []ProductResponse
	for _, p := range products {
		responses = append(responses, ToProductResponse(&p))
	}

	totalPages := (total + limit - 1) / limit

	return ProductListResponse{
		Products: responses,
		Pagination: PaginationResponse{
			CurrentPage:  page,
			TotalPages:   totalPages,
			TotalItems:   total,
			ItemsPerPage: limit,
		},
	}
}
