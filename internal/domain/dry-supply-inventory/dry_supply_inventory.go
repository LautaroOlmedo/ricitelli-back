package dry_supply_inventory

import (
	"errors"
	"fmt"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

// DrySupplyInventory tracks stock movements for a single dry supply item.
// Tricapa (three-layer) stock model:
//   - Physical Stock  = all IN movements - CONSUMED movements - ADJUSTED (negative)
//   - Committed Stock = COMMITTED movements - RELEASED movements
//   - Available Stock = Physical - Committed
type DrySupplyInventory struct {
	id          string
	drySupplyID string
	movements   []valueObject.DrySupplyMovement
}

func NewDrySupplyInventory(drySupplyID string) (DrySupplyInventory, error) {
	if drySupplyID == "" {
		return DrySupplyInventory{}, errors.New("drySupplyID cannot be empty")
	}
	return DrySupplyInventory{
		id:          uuid.New().String(),
		drySupplyID: drySupplyID,
		movements:   make([]valueObject.DrySupplyMovement, 0),
	}, nil
}

func (d *DrySupplyInventory) GetID() string          { return d.id }
func (d *DrySupplyInventory) GetDrySupplyID() string { return d.drySupplyID }
func (d *DrySupplyInventory) GetMovements() []valueObject.DrySupplyMovement {
	cp := make([]valueObject.DrySupplyMovement, len(d.movements))
	copy(cp, d.movements)
	return cp
}

// PhysicalStock = total IN - total CONSUMED ± ADJUSTED
func (d *DrySupplyInventory) PhysicalStock() int64 {
	var total int64
	for _, m := range d.movements {
		switch m.MovementType {
		case valueObject.DrySupplyIn:
			total += int64(m.Quantity)
		case valueObject.DrySupplyConsumed:
			total -= int64(m.Quantity)
		case valueObject.DrySupplyAdjusted:
			// Signed adjustment: positive quantity = add, stored as separate IN/OUT
			total += int64(m.Quantity)
		}
	}
	return total
}

// CommittedStock = total COMMITTED - total RELEASED - total CONSUMED
// (committed that have already been consumed are no longer "committed")
func (d *DrySupplyInventory) CommittedStock() int64 {
	var total int64
	for _, m := range d.movements {
		switch m.MovementType {
		case valueObject.DrySupplyCommitted:
			total += int64(m.Quantity)
		case valueObject.DrySupplyReleased:
			total -= int64(m.Quantity)
		case valueObject.DrySupplyConsumed:
			total -= int64(m.Quantity)
		}
	}
	if total < 0 {
		return 0
	}
	return total
}

// AvailableStock = Physical - Committed
func (d *DrySupplyInventory) AvailableStock() int64 {
	available := d.PhysicalStock() - d.CommittedStock()
	if available < 0 {
		return 0
	}
	return available
}

// AddStock records an incoming stock movement (purchase/receipt).
func (d *DrySupplyInventory) AddStock(quantity uint64, reference string) error {
	m, err := valueObject.NewDrySupplyMovement(valueObject.DrySupplyIn, quantity, reference)
	if err != nil {
		return err
	}
	d.movements = append(d.movements, m)
	return nil
}

// Commit reserves stock for a production order.
// Returns error if there is insufficient available stock.
func (d *DrySupplyInventory) Commit(quantity uint64, reference string) error {
	if int64(quantity) > d.AvailableStock() {
		return fmt.Errorf("insufficient available stock: need %d, have %d", quantity, d.AvailableStock())
	}
	m, err := valueObject.NewDrySupplyMovement(valueObject.DrySupplyCommitted, quantity, reference)
	if err != nil {
		return err
	}
	d.movements = append(d.movements, m)
	return nil
}

// Release cancels a previous commitment (e.g., production order cancelled).
func (d *DrySupplyInventory) Release(quantity uint64, reference string) error {
	m, err := valueObject.NewDrySupplyMovement(valueObject.DrySupplyReleased, quantity, reference)
	if err != nil {
		return err
	}
	d.movements = append(d.movements, m)
	return nil
}

// Consume records actual usage when production is completed.
func (d *DrySupplyInventory) Consume(quantity uint64, reference string) error {
	m, err := valueObject.NewDrySupplyMovement(valueObject.DrySupplyConsumed, quantity, reference)
	if err != nil {
		return err
	}
	d.movements = append(d.movements, m)
	return nil
}

// ReconstitueDrySupplyInventory reconstitutes a DrySupplyInventory from stored data.
func ReconstitueDrySupplyInventory(id, drySupplyID string, movements []valueObject.DrySupplyMovement) DrySupplyInventory {
	return DrySupplyInventory{
		id:          id,
		drySupplyID: drySupplyID,
		movements:   movements,
	}
}
