package product_inventory

import (
	"errors"
	"fmt"
	"time"

	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

// ProductInventory tracks stock movements for a wine product in two stages:
//   - UNDRESSED (Sin Vestir / SV): bottled but not yet labeled
//   - DRESSED   (Producto Terminado / PT): fully labeled and ready to dispatch
//
// Tricapa model per stage:
//   - Physical Stock  = STAGE_IN (or PRODUCED) - STAGE_OUT - DISPATCHED
//   - Committed Stock = RESERVED_FOR_SALE - RESERVATION_RELEASED - DISPATCHED
//   - Available Stock = Physical - Committed
type ProductInventory struct {
	id        string
	productID string
	sku       string
	movements []valueObject.ProductMovement
}

func NewProductInventory(productID, sku string) (ProductInventory, error) {
	if productID == "" || sku == "" {
		return ProductInventory{}, errors.New("product ID or SKU cannot be empty")
	}
	return ProductInventory{
		id:        uuid.New().String(),
		productID: productID,
		sku:       sku,
		movements: make([]valueObject.ProductMovement, 0),
	}, nil
}

func (p *ProductInventory) GetID() string        { return p.id }
func (p *ProductInventory) GetProductID() string { return p.productID }
func (p *ProductInventory) GetSku() string       { return p.sku }

func (p *ProductInventory) GetMovements() []valueObject.ProductMovement {
	cp := make([]valueObject.ProductMovement, len(p.movements))
	copy(cp, p.movements)
	return cp
}

// PhysicalUndressed = PRODUCED + STAGE_IN(undressed) - STAGE_OUT(undressed)
func (p *ProductInventory) PhysicalUndressed() int64 {
	var total int64
	for _, m := range p.movements {
		switch m.MovementType {
		case valueObject.ProductProduced:
			total += int64(m.Quantity)
		case valueObject.ProductStageIn:
			if m.Stage == valueObject.Undressed {
				total += int64(m.Quantity)
			}
		case valueObject.ProductStageOut:
			if m.Stage == valueObject.Undressed {
				total -= int64(m.Quantity)
			}
		}
	}
	return total
}

// PhysicalDressed = STAGE_IN(dressed) - DISPATCHED
func (p *ProductInventory) PhysicalDressed() int64 {
	var total int64
	for _, m := range p.movements {
		switch m.MovementType {
		case valueObject.ProductStageIn:
			if m.Stage == valueObject.Dressed {
				total += int64(m.Quantity)
			}
		case valueObject.ProductDispatched:
			total -= int64(m.Quantity)
		}
	}
	return total
}

// CommittedDressed = RESERVED_FOR_SALE - RESERVATION_RELEASED - DISPATCHED
func (p *ProductInventory) CommittedDressed() int64 {
	var total int64
	for _, m := range p.movements {
		if m.Stage != valueObject.Dressed {
			continue
		}
		switch m.MovementType {
		case valueObject.ProductReservedForSale:
			total += int64(m.Quantity)
		case valueObject.ProductReservationReleased:
			total -= int64(m.Quantity)
		case valueObject.ProductDispatched:
			total -= int64(m.Quantity)
		}
	}
	if total < 0 {
		return 0
	}
	return total
}

// AvailableDressed = PhysicalDressed - CommittedDressed
// This is the "ready to dispatch" stock.
func (p *ProductInventory) AvailableDressed() (int64, error) {
	available := p.PhysicalDressed() - p.CommittedDressed()
	if available < 0 {
		return 0, nil
	}
	return available, nil
}

// AvailableUndressed = PhysicalUndressed (all SV is available to be dressed)
func (p *ProductInventory) AvailableUndressed() int64 {
	return p.PhysicalUndressed()
}

// AddUndressed records new undressed (SV) wine arriving from production.
func (p *ProductInventory) AddUndressed(productionOrderID string, quantity uint64, userID string) error {
	m, err := valueObject.NewProductMovement(productionOrderID, valueObject.Undressed, valueObject.ProductStageIn, quantity, userID)
	if err != nil {
		return err
	}
	p.movements = append(p.movements, m)
	return nil
}

// ConvertSVtoPT converts undressed (SV) stock to dressed (PT) with a lot number.
// The lot number format is e.g. "L-081124-38-11".
func (p *ProductInventory) ConvertSVtoPT(referenceID string, quantity uint64, lotNumber string, userID string) error {
	if quantity == 0 {
		return errors.New("quantity must be greater than 0")
	}
	if lotNumber == "" {
		lotNumber = fmt.Sprintf("L-%s-%d", time.Now().Format("020106"), quantity)
	}
	if p.PhysicalUndressed() < int64(quantity) {
		return fmt.Errorf("insufficient undressed stock: need %d, have %d", quantity, p.PhysicalUndressed())
	}
	// Remove from SV
	outM, err := valueObject.NewProductMovementWithLot(referenceID, valueObject.Undressed, valueObject.ProductStageOut, quantity, lotNumber, userID)
	if err != nil {
		return err
	}
	// Add to PT
	inM, err := valueObject.NewProductMovementWithLot(referenceID, valueObject.Dressed, valueObject.ProductStageIn, quantity, lotNumber, userID)
	if err != nil {
		return err
	}
	p.movements = append(p.movements, outM, inM)
	return nil
}

// Reserve commits dressed stock for a sale order.
func (p *ProductInventory) Reserve(referenceID string, quantity uint64, userID string) error {
	m, err := valueObject.NewProductMovement(referenceID, valueObject.Dressed, valueObject.ProductReservedForSale, quantity, userID)
	if err != nil {
		return err
	}
	p.movements = append(p.movements, m)
	return nil
}

// ReleaseReservation cancels a previous dressed stock reservation (e.g., sale order cancelled).
func (p *ProductInventory) ReleaseReservation(referenceID string, quantity uint64, userID string) error {
	m, err := valueObject.NewProductMovement(referenceID, valueObject.Dressed, valueObject.ProductReservationReleased, quantity, userID)
	if err != nil {
		return err
	}
	p.movements = append(p.movements, m)
	return nil
}

// Dispatch records actual shipment, reducing physical and committed dressed stock.
func (p *ProductInventory) Dispatch(referenceID string, quantity uint64, userID string) error {
	m, err := valueObject.NewProductMovement(referenceID, valueObject.Dressed, valueObject.ProductDispatched, quantity, userID)
	if err != nil {
		return err
	}
	p.movements = append(p.movements, m)
	return nil
}

func (p *ProductInventory) ValidateMovements(productID string) error {
	if len(p.movements) == 0 {
		return fmt.Errorf("product %s has no inventory movements", productID)
	}
	return nil
}

// ReconstitueProductInventory reconstitutes a ProductInventory from stored data.
func ReconstitueProductInventory(id, productID, sku string, movements []valueObject.ProductMovement) ProductInventory {
	return ProductInventory{
		id:        id,
		productID: productID,
		sku:       sku,
		movements: movements,
	}
}
