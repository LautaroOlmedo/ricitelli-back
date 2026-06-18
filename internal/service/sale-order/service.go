package sale_order

import (
	"context"

	"ricitelli-back/internal/domain/remittance"
	sale_order "ricitelli-back/internal/domain/sale-order"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
)

type SaleOrderStorage interface {
	CreateSaleOrder(ctx context.Context, params sale_order.NewSaleOrderParams) (sale_order.SaleOrder, error)
	GetSaleOrderByID(ctx context.Context, id string) (*sale_order.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order.SaleOrder, error)
	GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order.SaleOrder, error)
	UpdateSaleOrderStatus(ctx context.Context, id string, status sale_order.Status) (*sale_order.SaleOrder, error)
	ListSalesInvoices(ctx context.Context, filter sales_administration.SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error)
	ListRemittances(ctx context.Context, filter sales_administration.RemittanceFilter) ([]remittance.Remittance, error)
}

type Service struct {
	Storage SaleOrderStorage
}

func NewSaleOrderService(saleOrderStorage SaleOrderStorage) *Service {
	return &Service{Storage: saleOrderStorage}
}
