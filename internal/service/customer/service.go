package customer

import (
	"context"

	customer_domain "ricitelli-back/internal/domain/customer"
	sale_order "ricitelli-back/internal/domain/sale-order"
	valueObject "ricitelli-back/internal/value-object"
)

type CustomerStorage interface {
	CreateCustomer(ctx context.Context, params customer_domain.NewCustomerParams) (customer_domain.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error)
	GetCustomers(ctx context.Context) ([]customer_domain.Customer, error)
	DeactivateCustomer(ctx context.Context, id string) (*customer_domain.Customer, error)
}

type SaleOrderStorage interface {
	CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
}

// PlaceOrderParams holds the order details provided by the caller.
// The Market field on the resulting SaleOrder is inferred from the customer's MarketType
// unless explicitly overridden.
type PlaceOrderParams struct {
	CustomerID         string
	Items              []valueObject.SaleOrderItem
	Currency           sale_order.Currency
	DestinationCountry string
	SaleType           sale_order.SaleType
}

type Service struct {
	Storage          CustomerStorage
	SaleOrderStorage SaleOrderStorage
}

func NewCustomerService(storage CustomerStorage, saleOrderStorage SaleOrderStorage) *Service {
	return &Service{
		Storage:          storage,
		SaleOrderStorage: saleOrderStorage,
	}
}
