package application_service

import (
	"context"

	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
	"ricitelli-back/internal/domain/product"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	valueObject "ricitelli-back/internal/value-object"
)

//go:generate mockgen -source=service.go -destination=././mocks/sale_order_service_mock.go -package=mocks
type SaleOrderService interface {
	CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
	UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error)
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

type ProductInventoryService interface {
	CreateProductInventory(productID, sku string) error
	GetProductInventory(productID string) (*product_inventory.ProductInventory, error)
	SaveProductInventory(inv product_inventory.ProductInventory) error
}

type DrySupplyService interface {
	CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string) (*dry_supply.DrySupply, error)
	GetDrySupplyByID(ctx context.Context, id string) (*dry_supply.DrySupply, error)
	GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error)
	AddStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	CommitStock(ctx context.Context, drySupplyID string, quantity uint64, productionOrderID string) error
	ReleaseStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	ConsumeStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error)
}

type Service struct {
	SaleOrderService        SaleOrderService
	ProductionOrderService  ProductionOrderService
	ProductService          ProductService
	ProductInventoryService ProductInventoryService
	DrySupplyService        DrySupplyService
}

func NewApplicationService(
	saleOrderService SaleOrderService,
	productionOrderService ProductionOrderService,
	productService ProductService,
	productInventoryService ProductInventoryService,
	drySupplyService DrySupplyService,
) Service {
	return Service{
		SaleOrderService:        saleOrderService,
		ProductionOrderService:  productionOrderService,
		ProductService:          productService,
		ProductInventoryService: productInventoryService,
		DrySupplyService:        drySupplyService,
	}
}
