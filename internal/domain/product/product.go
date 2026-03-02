package product

import (
	"errors"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/google/uuid"
)

var (
	ErrInvalidName             = errors.New("invalid name")
	ErrInvalidQuantityOfSupply = errors.New("invalid quantity of bill of dry supply item")
)

// Product represents a manufacturable item defined by its identity, name, and its bill of materials (BODS).
type Product struct {
	id   string
	name string
	bom  []valueObject.BillOfDrySupply
}

func NewProduct(name string, bom []valueObject.BillOfDrySupply) (Product, error) {
	if name == "" {
		// handle name already exists case
		return Product{}, ErrInvalidName
	}
	if len(bom) == 0 {
		return Product{}, ErrInvalidQuantityOfSupply
	}

	return Product{
		id:   uuid.New().String(),
		name: name,
		bom:  bom,
	}, nil
}

func (p *Product) GetID() string {
	return p.id
}

func (p *Product) GetName() string {
	return p.name
}

func (p *Product) CalculateRequirements(qty uint64) []valueObject.MaterialRequirement {
	var result []valueObject.MaterialRequirement

	for _, item := range p.bom {
		total := item.QuantityPerUnit * qty

		result = append(result, valueObject.MaterialRequirement{
			DrySupplyID: item.DrySupplyID,
			Quantity:    total,
		})
	}

	return result
}
