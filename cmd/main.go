package main

import (
	"log"

	"github.com/ifortepay/ApiBackendTest/internal/config"
	routes "github.com/ifortepay/ApiBackendTest/internal/route"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg := config.Load()

	container, err := config.BuildContainer(cfg)
	if err != nil {
		log.Fatalf("failed build container: %v", err)
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	routes.Register(e, &routes.Handler{
		CheckoutHandler: container.CheckoutHandler,
		ProductHandler:  container.ProductHandler,
		CartHandler:     container.CartHandler,
		OrderHandler:    container.OrderHandler,
	})

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
