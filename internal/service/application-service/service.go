package application_service

import (
	"context"

	"ricitelli-back/internal/auth"
	customer_domain "ricitelli-back/internal/domain/customer"
	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
	"ricitelli-back/internal/domain/product"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	production_order "ricitelli-back/internal/domain/production-order"
	sale_order "ricitelli-back/internal/domain/sale-order"
	"ricitelli-back/internal/entities"
	customer_svc "ricitelli-back/internal/service/customer"
	valueObject "ricitelli-back/internal/value-object"
)

func userIDFromCtx(ctx context.Context) string {
	if claims, ok := auth.ClaimsFromContext(ctx); ok {
		return claims.UserID
	}
	return "system"
}

type CustomerService interface {
	CreateCustomer(ctx context.Context, params customer_domain.NewCustomerParams) (customer_domain.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error)
	GetCustomers(ctx context.Context) ([]customer_domain.Customer, error)
	DeactivateCustomer(ctx context.Context, id string) (*customer_domain.Customer, error)
	PlaceOrder(ctx context.Context, params customer_svc.PlaceOrderParams) (sale_order.SaleOrder, error)
	GetOrdersByCustomer(ctx context.Context, customerID string) ([]sale_order.SaleOrder, error)
}

//go:generate mockgen -source=service.go -destination=././mocks/sale_order_service_mock.go -package=mocks
type SaleOrderService interface {
	CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
	UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error)
}

//go:generate mockgen -source=service.go -destination=././mocks/production_order_service_mock.go -package=mocks
type ProductionOrderService interface {
	CreateProductionOrder(ctx context.Context, salesOrderID string, items []entities.ProductionItem) (production_order.ProductionOrder, error)
	GetProductionOrderByID(ctx context.Context, id string) (*production_order.ProductionOrder, error)
	GetProductionOrders(ctx context.Context) ([]production_order.ProductionOrder, error)
	GetProductionOrdersBySaleOrder(ctx context.Context, saleOrderID string) ([]production_order.ProductionOrder, error)
	UpdateProductionOrderStatus(ctx context.Context, id string, newStatus production_order.Status) (*production_order.ProductionOrder, error)
}

//go:generate mockgen -source=service.go -destination=././mocks/product_service_mock.go -package=mocks
type ProductService interface {
	CreateProduct(ctx context.Context, name string, bods []valueObject.BillOfDrySupply) (*product.Product, error)
	GetProductByID(ctx context.Context, id string) (*product.Product, error)
	GetProducts(ctx context.Context) ([]product.Product, error)
	UpdateProduct(ctx context.Context, id, name string, bods []valueObject.BillOfDrySupply) error
	SetProductImage(ctx context.Context, id, imageURL string) error
	GetProductImage(ctx context.Context, id string) (string, error)
}

type ProductInventoryService interface {
	CreateProductInventory(productID, sku string) error
	GetProductInventory(productID string) (*product_inventory.ProductInventory, error)
	SaveProductInventory(inv product_inventory.ProductInventory) error
}

type DrySupplyService interface {
	CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string, reorderPoint int) (*dry_supply.DrySupply, error)
	GetDrySupplyByID(ctx context.Context, id string) (*dry_supply.DrySupply, error)
	GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error)
	AddStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	CommitStock(ctx context.Context, drySupplyID string, quantity uint64, productionOrderID string) error
	ReleaseStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	ConsumeStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error
	GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error)
}

type Service struct {
	CustomerService         CustomerService
	SaleOrderService        SaleOrderService
	ProductionOrderService  ProductionOrderService
	ProductService          ProductService
	ProductInventoryService ProductInventoryService
	DrySupplyService        DrySupplyService
}

func NewApplicationService(
	customerService CustomerService,
	saleOrderService SaleOrderService,
	productionOrderService ProductionOrderService,
	productService ProductService,
	productInventoryService ProductInventoryService,
	drySupplyService DrySupplyService,
) Service {
	return Service{
		CustomerService:         customerService,
		SaleOrderService:        saleOrderService,
		ProductionOrderService:  productionOrderService,
		ProductService:          productService,
		ProductInventoryService: productInventoryService,
		DrySupplyService:        drySupplyService,
	}
}
