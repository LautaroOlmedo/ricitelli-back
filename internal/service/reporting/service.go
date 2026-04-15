package reporting

import (
	"context"
	"fmt"
	"time"

	customer_domain "ricitelli-back/internal/domain/customer"
	dry_supply_domain "ricitelli-back/internal/domain/dry-supply"
	product_domain "ricitelli-back/internal/domain/product"
	production_order_domain "ricitelli-back/internal/domain/production-order"
	sale_order_domain "ricitelli-back/internal/domain/sale-order"
	inventory_svc "ricitelli-back/internal/service/inventory"
	"ricitelli-back/internal/service/reporting/storage"
)

// Service-level interfaces so reporting depends only on abstractions.

type SaleOrderSource interface {
	GetSaleOrdersByDateRange(ctx context.Context, from, to string) ([]sale_order_domain.SaleOrder, error)
	GetSaleOrders(ctx context.Context) ([]sale_order_domain.SaleOrder, error)
}

type ProductionOrderSource interface {
	GetProductionOrders(ctx context.Context) ([]production_order_domain.ProductionOrder, error)
}

type ProductSource interface {
	GetProducts(ctx context.Context) ([]product_domain.Product, error)
	GetProductByID(ctx context.Context, id string) (*product_domain.Product, error)
}

type DrySupplySource interface {
	GetDrySupplies(ctx context.Context) ([]dry_supply_domain.DrySupply, error)
	GetDrySupplyByID(ctx context.Context, id string) (*dry_supply_domain.DrySupply, error)
}

type CustomerSource interface {
	GetCustomers(ctx context.Context) ([]customer_domain.Customer, error)
	GetCustomerByID(ctx context.Context, id string) (*customer_domain.Customer, error)
}

type InventorySource interface {
	GetMovements(ctx context.Context, filter inventory_svc.MovementFilter) (*inventory_svc.MovementsResult, error)
	GetInventoryReport(ctx context.Context) (*inventory_svc.InventoryReport, error)
	GetLowStockAlerts(ctx context.Context) ([]inventory_svc.DrySupplyAlert, error)
}

// Service orchestrates report generation.
type Service struct {
	Sales        SaleOrderSource
	Production   ProductionOrderSource
	Products     ProductSource
	DrySupplies  DrySupplySource
	Customers    CustomerSource
	Inventory    InventorySource
	Storage      *storage.Store
	PublicURLBase string
}

func NewService(
	sales SaleOrderSource,
	production ProductionOrderSource,
	products ProductSource,
	drySupplies DrySupplySource,
	customers CustomerSource,
	inventory InventorySource,
	store *storage.Store,
	publicURLBase string,
) *Service {
	return &Service{
		Sales:         sales,
		Production:    production,
		Products:      products,
		DrySupplies:   drySupplies,
		Customers:     customers,
		Inventory:     inventory,
		Storage:       store,
		PublicURLBase: publicURLBase,
	}
}

// Result is the metadata returned after generating a report.
type Result struct {
	ID          string
	Type        ReportType
	Filename    string
	DownloadURL string
	GeneratedAt string
	FileSize    int64
	FromDate    string
	ToDate      string
}

func (s *Service) buildResult(m *storage.Metadata) *Result {
	return &Result{
		ID:          m.ID,
		Type:        ReportType(m.Type),
		Filename:    m.Filename,
		DownloadURL: fmt.Sprintf("%s/%s", s.PublicURLBase, m.ID),
		GeneratedAt: m.GeneratedAt.UTC().Format(time.RFC3339),
		FileSize:    m.FileSize,
		FromDate:    m.FromDate,
		ToDate:      m.ToDate,
	}
}

// ListReports returns stored report metadata.
func (s *Service) ListReports(ctx context.Context, typeFilter string, page, pageSize int) ([]*Result, int, error) {
	if pageSize <= 0 {
		pageSize = 50
	}
	if page <= 0 {
		page = 1
	}
	all := s.Storage.List(typeFilter)
	total := len(all)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	out := make([]*Result, 0, end-start)
	for _, m := range all[start:end] {
		out = append(out, s.buildResult(m))
	}
	return out, total, nil
}
