package handler

import (
	"net/http"

	"github.com/ifortepay/ApiBackendTest/internal/dto"
	"github.com/ifortepay/ApiBackendTest/internal/model"
	"github.com/ifortepay/ApiBackendTest/internal/pkg/utils"
	"github.com/ifortepay/ApiBackendTest/internal/service"
	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	productService service.ProductService
}

func NewProductHandler(productService service.ProductService) *ProductHandler {
	return &ProductHandler{
		productService: productService,
	}
}

// GetProducts handles GET /api/products
func (h *ProductHandler) GetProducts(c echo.Context) error {
	var req dto.GetProductsRequest
	if err := c.Bind(&req); err != nil {
		return utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters")
	}

	filter := model.ProductFilter{
		Search:   req.Search,
		MinPrice: req.MinPrice,
		MaxPrice: req.MaxPrice,
		SKU:      req.SKU,
		Page:     req.Page,
		Limit:    req.Limit,
	}

	products, total, err := h.productService.GetProducts(c.Request().Context(), filter)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get products")
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToProductListResponse(products, total, filter.Page, filter.Limit))
}

// GetProductBySKU handles GET /api/products/:sku
func (h *ProductHandler) GetProductBySKU(c echo.Context) error {
	sku := c.Param("sku")
	if sku == "" {
		return utils.ErrorResponse(c, http.StatusBadRequest, "SKU is required")
	}

	product, err := h.productService.GetProductBySKU(c.Request().Context(), sku)
	if err != nil {
		return utils.ErrorResponse(c, http.StatusNotFound, "Product not found")
	}

	return utils.SuccessResponse(c, http.StatusOK, dto.ToProductResponse(product))
}
