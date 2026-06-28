package utils

import (
	"github.com/labstack/echo/v4"
)

type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func SuccessResponse(c echo.Context, statusCode int, data interface{}) error {
	return c.JSON(statusCode, APIResponse{
		Status: "success",
		Data:   data,
	})
}

func ErrorResponse(c echo.Context, statusCode int, message string) error {
	return c.JSON(statusCode, APIResponse{
		Status:  "error",
		Message: message,
	})
}
