package product_inventory

import (
	"errors"
	"fmt"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

type ProductInventory struct {
	id        string
	productID string
	sku       string
	movements []valueObject.ProductMovement
}

func NewProductInventory(productID, sku string) (ProductInventory, error) {
	if productID == "" || sku == "" {
		return ProductInventory{}, errors.New(fmt.Sprintf("Product ID or SKU cannot be empty"))
	}
	return ProductInventory{
		id:        uuid.New().String(),
		productID: productID,
		sku:       sku,
		movements: make([]valueObject.ProductMovement, 0),
	}, nil
}

func (p *ProductInventory) GetID() string {
	return p.id
}

func (p *ProductInventory) GetProductID() string {
	return p.productID
}

func (p *ProductInventory) GetSku() string {
	return p.sku
}

func (p *ProductInventory) GetMovements() []valueObject.ProductMovement {
	itemsCopy := make([]valueObject.ProductMovement, len(p.movements))
	copy(itemsCopy, p.movements)
	return itemsCopy
}

func (p *ProductInventory) AvailableUndressed() int64 {
	return 1
}

// AvailableDressed allows you to create an order without an order production. It means stock available and ready to dispatch
func (p *ProductInventory) AvailableDressed() (int64, error) {

	var total int64

	for _, m := range p.movements {

		if m.Stage != valueObject.Dressed {
			continue
		}
		switch m.MovementType {

		case valueObject.ProductStageIn:
			total += int64(m.Quantity)

		case valueObject.ProductReservedForSale:
			total -= int64(m.Quantity)

		case valueObject.ProductReservationReleased:
			total += int64(m.Quantity)

		case valueObject.ProductDispatched:
			total -= int64(m.Quantity)

			/*case valueObject.ProductStockAdjusted:
				total += int64(m.Quantity)
			}*/
		}
		fmt.Println("total:", total)
	}

	return total, nil
}

func (p *ProductInventory) Reserve(referenceID string, quantity uint64) error {
	newMovement, err := valueObject.NewProductMovement(referenceID, valueObject.Dressed, valueObject.ProductReservedForSale, quantity)
	if err != nil {
		return err
	}

	p.movements = append(p.movements, newMovement)
	return nil
}

func (p *ProductInventory) ValidateMovements(productID string) error {
	if len(p.movements) == 0 {
		fmt.Printf("[ERROR] Product %s has no inventory movements", productID)
		return errors.New("no inventory movements")
	}
	return nil
}

/*func (p *ProductInventory) Reserve(productID, referenceID string, stage valueObject.Stage, quantity uint64) error {
	inventoryMovement, err := valueObject.NewProductMovement(referenceID, stage, valueObject.ProductReservedForSale, quantity)
	if err != nil {
		return err
	}
	// mu.Lock()
	p.movements = append(p.movements, inventoryMovement)
	return nil
}
func (p *ProductInventory) IncreaseInProduction(productID, referenceID string, qty uint64) error {
	if err := p.validateMovements(productID); err != nil {
		return err
	}
	for _, mov := range p.movements {
		if mov.Reference == referenceID {
			mov.MovementType = "IN PRODUCTION"
			mov.Quantity = qty

		} else {
			fmt.Printf("[ERROR] inventory movement for reference %s not found", referenceID)
			return errors.New("reference does not have an inventory movement")
		}
	}

	return nil
}

func (p *ProductInventory) FromProductionToAvailable(productID, referenceID string) error {
	if err := p.validateMovements(productID); err != nil {
		return err
	}
	for _, mov := range p.movements {
		if mov.Reference == referenceID {
			mov.MovementType = "AVAILABLE"
		} else {
			fmt.Printf("[ERROR] inventory movement for reference %s not found", referenceID)
			return errors.New("reference does not have an inventory movement")
		}
	}
	return nil
}*/
