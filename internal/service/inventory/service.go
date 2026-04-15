package inventory

import (
	"context"

	"ricitelli-back/internal/auth"
	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	productdomain "ricitelli-back/internal/domain/product"
)

// ProductTricapa holds tricapa metrics + undressed stock for a wine product.
type ProductTricapa struct {
	ProductID      string
	ProductName    string
	SKU            string
	UndressedStock int64 // SV - Sin Vestir (physical)
	DressedPhysical int64 // PT physical
	DressedCommitted int64 // PT committed for sale orders
	DressedAvailable int64 // PT available = physical - committed
}

// DrySupplyAlert flags a dry supply whose available stock is below required.
type DrySupplyAlert struct {
	DrySupplyID   string
	Code          string
	Name          string
	Physical      int64
	Committed     int64
	Available     int64
	IsLow         bool // available < lowStockThreshold
}

// InventoryReport consolidates all tricapa metrics.
type InventoryReport struct {
	Products         []ProductTricapa
	DrySupplyAlerts  []DrySupplyAlert
}

type ProductStorage interface {
	GetProducts(ctx context.Context) ([]productdomain.Product, error)
	GetProductByID(ctx context.Context, id string) (*productdomain.Product, error)
}

type ProductInventoryStorage interface {
	GetProductInventory(productID string) (*product_inventory.ProductInventory, error)
	SaveProductInventory(inv product_inventory.ProductInventory) error
}

type DrySupplyStorage interface {
	GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error)
	GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error)
	GetDailyLotCount(ctx context.Context) (int, error)
}

type MovementStorage interface {
	QueryMovements(ctx context.Context, filter MovementFilter) (*MovementsResult, error)
}

// MovementFilter holds query parameters for the audit trail.
type MovementFilter struct {
	FromDate     string // RFC3339
	ToDate       string // RFC3339
	UserID       string
	MovementType string
	ProductID    string
	DrySupplyID  string
	Category     string // "PRODUCT", "DRY_SUPPLY", or "" for all
	Page         int
	PageSize     int
}

// MovementEntry is a denormalized audit row returned by QueryMovements.
type MovementEntry struct {
	MovementType string
	Quantity     uint64
	Reference    string
	Stage        string
	LotNumber    string
	UserID       string
	CreatedAt    string
	ItemName     string
	Category     string // "PRODUCT" or "DRY_SUPPLY"
}

// MovementsResult wraps the paginated query result.
type MovementsResult struct {
	Movements  []MovementEntry
	TotalCount int
}

const defaultLowStockThreshold = 500

func userIDFromContext(ctx context.Context) string {
	if claims, ok := auth.ClaimsFromContext(ctx); ok {
		return claims.UserID
	}
	return "system"
}

type Service struct {
	productStorage   ProductStorage
	productInventory ProductInventoryStorage
	drySupplyStorage DrySupplyStorage
	movementStorage  MovementStorage
}

func NewInventoryService(
	productStorage ProductStorage,
	productInventory ProductInventoryStorage,
	drySupplyStorage DrySupplyStorage,
	movementStorage MovementStorage,
) *Service {
	return &Service{
		productStorage:   productStorage,
		productInventory: productInventory,
		drySupplyStorage: drySupplyStorage,
		movementStorage:  movementStorage,
	}
}

// GetInventoryReport returns a full tricapa snapshot for all products and dry supplies.
func (s *Service) GetInventoryReport(ctx context.Context) (*InventoryReport, error) {
	products, err := s.productStorage.GetProducts(ctx)
	if err != nil {
		return nil, err
	}

	var productTricapas []ProductTricapa
	for _, p := range products {
		inv, err := s.productInventory.GetProductInventory(p.GetID())
		if err != nil {
			// No inventory record yet — report zeros
			productTricapas = append(productTricapas, ProductTricapa{
				ProductID:   p.GetID(),
				ProductName: p.GetName(),
			})
			continue
		}
		available, _ := inv.AvailableDressed()
		productTricapas = append(productTricapas, ProductTricapa{
			ProductID:        p.GetID(),
			ProductName:      p.GetName(),
			SKU:              inv.GetSku(),
			UndressedStock:   inv.AvailableUndressed(),
			DressedPhysical:  inv.PhysicalDressed(),
			DressedCommitted: inv.CommittedDressed(),
			DressedAvailable: available,
		})
	}

	drySupplies, err := s.drySupplyStorage.GetDrySupplies(ctx)
	if err != nil {
		return nil, err
	}

	var alerts []DrySupplyAlert
	for _, ds := range drySupplies {
		inv, err := s.drySupplyStorage.GetDrySupplyInventory(ctx, ds.GetID())
		if err != nil {
			alerts = append(alerts, DrySupplyAlert{
				DrySupplyID: ds.GetID(),
				Code:        ds.GetCode(),
				Name:        ds.GetName(),
				IsLow:       true,
			})
			continue
		}
		available := inv.AvailableStock()
		alerts = append(alerts, DrySupplyAlert{
			DrySupplyID: ds.GetID(),
			Code:        ds.GetCode(),
			Name:        ds.GetName(),
			Physical:    inv.PhysicalStock(),
			Committed:   inv.CommittedStock(),
			Available:   available,
			IsLow:       ds.GetReorderPoint() > 0 && available < int64(ds.GetReorderPoint()),
		})
	}

	return &InventoryReport{
		Products:        productTricapas,
		DrySupplyAlerts: alerts,
	}, nil
}

// GetProductTricapa returns tricapa metrics for a single product.
func (s *Service) GetProductTricapa(ctx context.Context, productID string) (*ProductTricapa, error) {
	p, err := s.productStorage.GetProductByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	inv, err := s.productInventory.GetProductInventory(p.GetID())
	if err != nil {
		return &ProductTricapa{ProductID: p.GetID(), ProductName: p.GetName()}, nil
	}
	available, _ := inv.AvailableDressed()
	return &ProductTricapa{
		ProductID:        p.GetID(),
		ProductName:      p.GetName(),
		SKU:              inv.GetSku(),
		UndressedStock:   inv.AvailableUndressed(),
		DressedPhysical:  inv.PhysicalDressed(),
		DressedCommitted: inv.CommittedDressed(),
		DressedAvailable: available,
	}, nil
}

// ConvertSVtoPT converts undressed (SV) wine to dressed (PT) and assigns a lot number.
// If lotNumber is empty, one is generated automatically in the format L-DDMMYY-NNN-XX.
func (s *Service) ConvertSVtoPT(ctx context.Context, productID string, quantity uint64, lotNumber string) error {
	inv, err := s.productInventory.GetProductInventory(productID)
	if err != nil {
		return err
	}
	if lotNumber == "" {
		dailyCount, _ := s.drySupplyStorage.GetDailyLotCount(ctx)
		lotNumber = GenerateLotNumber(productID, dailyCount+1)
	}
	if err := inv.ConvertSVtoPT("manual-conversion", quantity, lotNumber, userIDFromContext(ctx)); err != nil {
		return err
	}
	return s.productInventory.SaveProductInventory(*inv)
}

// AddUndressedStock ingests new SV (sin vestir) bottles into a product's inventory.
func (s *Service) AddUndressedStock(ctx context.Context, productID string, quantity uint64, reference string) error {
	inv, err := s.productInventory.GetProductInventory(productID)
	if err != nil {
		return err
	}
	if reference == "" {
		reference = "manual"
	}
	if err := inv.AddUndressed(reference, quantity, userIDFromContext(ctx)); err != nil {
		return err
	}
	return s.productInventory.SaveProductInventory(*inv)
}

// GetLowStockAlerts returns only the dry supply items below the stock threshold.
func (s *Service) GetLowStockAlerts(ctx context.Context) ([]DrySupplyAlert, error) {
	report, err := s.GetInventoryReport(ctx)
	if err != nil {
		return nil, err
	}
	var alerts []DrySupplyAlert
	for _, a := range report.DrySupplyAlerts {
		if a.IsLow {
			alerts = append(alerts, a)
		}
	}
	return alerts, nil
}

// GetMovements queries the audit trail with optional filters.
func (s *Service) GetMovements(ctx context.Context, filter MovementFilter) (*MovementsResult, error) {
	if filter.PageSize <= 0 {
		filter.PageSize = 50
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	return s.movementStorage.QueryMovements(ctx, filter)
}
