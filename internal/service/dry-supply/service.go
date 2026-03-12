package dry_supply

import (
	"context"

	dry_supply "ricitelli-back/internal/domain/dry-supply"
	dry_supply_inventory "ricitelli-back/internal/domain/dry-supply-inventory"
)

// StockTricapa holds the three-layer stock metrics for a dry supply.
type StockTricapa struct {
	DrySupplyID   string
	DrySupplyCode string
	DrySupplyName string
	Physical      int64
	Committed     int64
	Available     int64
}

type DrySupplyStorage interface {
	CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string) (*dry_supply.DrySupply, error)
	GetDrySupplyByID(ctx context.Context, id string) (*dry_supply.DrySupply, error)
	GetDrySupplyByCode(ctx context.Context, code string) (*dry_supply.DrySupply, error)
	GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error)
	UpdateDrySupply(ctx context.Context, id, name string, reorderPoint int) error
	GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error)
	SaveDrySupplyInventory(ctx context.Context, inv dry_supply_inventory.DrySupplyInventory) error
	GetDailyLotCount(ctx context.Context) (int, error)
}

type Service struct {
	storage DrySupplyStorage
}

func NewDrySupplyService(storage DrySupplyStorage) *Service {
	return &Service{storage: storage}
}

func (s *Service) CreateDrySupply(ctx context.Context, code, name string, category dry_supply.Category, unit string) (*dry_supply.DrySupply, error) {
	return s.storage.CreateDrySupply(ctx, code, name, category, unit)
}

func (s *Service) GetDrySupplyByID(ctx context.Context, id string) (*dry_supply.DrySupply, error) {
	return s.storage.GetDrySupplyByID(ctx, id)
}

func (s *Service) GetDrySupplies(ctx context.Context) ([]dry_supply.DrySupply, error) {
	return s.storage.GetDrySupplies(ctx)
}

// AddStock records incoming stock for a dry supply (purchase/receipt).
func (s *Service) AddStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error {
	inv, err := s.storage.GetDrySupplyInventory(ctx, drySupplyID)
	if err != nil {
		return err
	}
	if err := inv.AddStock(quantity, reference); err != nil {
		return err
	}
	return s.storage.SaveDrySupplyInventory(ctx, *inv)
}

// GetStockTricapa returns the three-layer stock metrics for a dry supply.
func (s *Service) GetStockTricapa(ctx context.Context, drySupplyID string) (*StockTricapa, error) {
	ds, err := s.storage.GetDrySupplyByID(ctx, drySupplyID)
	if err != nil {
		return nil, err
	}
	inv, err := s.storage.GetDrySupplyInventory(ctx, drySupplyID)
	if err != nil {
		return nil, err
	}
	return &StockTricapa{
		DrySupplyID:   ds.GetID(),
		DrySupplyCode: ds.GetCode(),
		DrySupplyName: ds.GetName(),
		Physical:      inv.PhysicalStock(),
		Committed:     inv.CommittedStock(),
		Available:     inv.AvailableStock(),
	}, nil
}

// CommitStock reserves stock for a production order.
func (s *Service) CommitStock(ctx context.Context, drySupplyID string, quantity uint64, productionOrderID string) error {
	inv, err := s.storage.GetDrySupplyInventory(ctx, drySupplyID)
	if err != nil {
		return err
	}
	if err := inv.Commit(quantity, productionOrderID); err != nil {
		return err
	}
	return s.storage.SaveDrySupplyInventory(ctx, *inv)
}

// ReleaseStock cancels a dry supply commitment.
func (s *Service) ReleaseStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error {
	inv, err := s.storage.GetDrySupplyInventory(ctx, drySupplyID)
	if err != nil {
		return err
	}
	if err := inv.Release(quantity, reference); err != nil {
		return err
	}
	return s.storage.SaveDrySupplyInventory(ctx, *inv)
}

// ConsumeStock records actual consumption when production is complete.
func (s *Service) ConsumeStock(ctx context.Context, drySupplyID string, quantity uint64, reference string) error {
	inv, err := s.storage.GetDrySupplyInventory(ctx, drySupplyID)
	if err != nil {
		return err
	}
	if err := inv.Consume(quantity, reference); err != nil {
		return err
	}
	return s.storage.SaveDrySupplyInventory(ctx, *inv)
}

// GetDrySupplyInventory returns the inventory record for a dry supply.
func (s *Service) GetDrySupplyInventory(ctx context.Context, drySupplyID string) (*dry_supply_inventory.DrySupplyInventory, error) {
	return s.storage.GetDrySupplyInventory(ctx, drySupplyID)
}

// GetDailyLotCount returns the count of lot numbers generated today.
func (s *Service) GetDailyLotCount(ctx context.Context) (int, error) {
	return s.storage.GetDailyLotCount(ctx)
}
