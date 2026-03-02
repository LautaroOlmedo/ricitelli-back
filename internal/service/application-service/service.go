package application_service

import (
	"context"
	"ricitelli-back/internal/domain/product"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
)

//go:generate mockgen -source=service.go -destination=././mocks/sale_order_service_mock.go -package=mocks
type SaleOrderService interface {
	CreateSaleOrder(ctx context.Context, customerID string, items []valueObject.SaleOrderItem) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
}

//go:generate mockgen -source=service.go -destination=././mocks/production_order_service_mock.go -package=mocks
type ProductionOrderService interface {
	CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) error
	GetProductionOrderByID(ctx context.Context, id string) (*production_order.ProductionOrder, error)
	GetProductionOrders(ctx context.Context) ([]production_order.ProductionOrder, error)
}

//go:generate mockgen -source=service.go -destination=././mocks/product_service_mock.go -package=mocks
type ProductService interface {
	CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) error
	GetProductByID(ctx context.Context, id string) (*product.Product, error)
	GetProducts(ctx context.Context) ([]product.Product, error)
}

type Service struct {
	SaleOrderService       SaleOrderService
	ProductionOrderService ProductionOrderService
	ProductService         ProductService
}

func NewApplicationService(productService ProductService, productionOrderService ProductionOrderService, saleOrderService SaleOrderService) Service {
	return Service{
		SaleOrderService:       saleOrderService,
		ProductionOrderService: productionOrderService,
		ProductService:         productService,
	}
}
