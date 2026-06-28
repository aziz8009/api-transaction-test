package config

import (
	"github.com/ifortepay/ApiBackendTest/internal/database"
	"github.com/ifortepay/ApiBackendTest/internal/handler"
	"github.com/ifortepay/ApiBackendTest/internal/repository"
	"github.com/ifortepay/ApiBackendTest/internal/service"
)

var newPostgresConnection = database.NewPostgresConnection

type Container struct {
	CheckoutHandler *handler.CheckoutHandler
	ProductHandler  *handler.ProductHandler
	CartHandler     *handler.CartHandler
	OrderHandler    *handler.OrderHandler
}

func BuildContainer(cfg *Config) (*Container, error) {
	db, err := newPostgresConnection(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	txManager := database.NewTransactionManager(db.GetDB())

	cartRepo := repository.NewCartRepository()
	productRepo := repository.NewProductRepository()
	orderRepo := repository.NewOrderRepository()
	promoRepo := repository.NewPromotionRepository()

	productService := service.NewProductService(productRepo, db)
	cartService := service.NewCartService(cartRepo, productRepo, txManager, db)
	promotionService := service.NewPromotionService(promoRepo, productRepo, db)
	checkoutService := service.NewCheckoutService(cartRepo, productRepo, orderRepo, promotionService, txManager, db)

	return &Container{
		ProductHandler:  handler.NewProductHandler(productService),
		CartHandler:     handler.NewCartHandler(cartService),
		CheckoutHandler: handler.NewCheckoutHandler(checkoutService),
		OrderHandler:    handler.NewOrderHandler(checkoutService),
	}, nil
}
