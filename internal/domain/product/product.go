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
	id     string
	name   string
	bods   []valueObject.BillOfDrySupply
	active bool
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
		id:     uuid.New().String(),
		name:   name,
		bods:   bom,
		active: true,
	}, nil
}

func NewProductWithID(id string, name string, bods []valueObject.BillOfDrySupply) (Product, error) {
	if id == "" {
		return Product{}, errors.New("invalid id")
	}
	if name == "" {
		return Product{}, ErrInvalidName
	}
	if len(bods) == 0 {
		return Product{}, ErrInvalidQuantityOfSupply
	}

	return Product{
		id:     id,
		name:   name,
		bods:   bods,
		active: true,
	}, nil
}

func (p *Product) GetID() string {
	return p.id
}

func (p *Product) GetName() string {
	return p.name
}

func (p *Product) GetBODS() []valueObject.BillOfDrySupply {
	itemsCopy := make([]valueObject.BillOfDrySupply, len(p.bods))
	copy(itemsCopy, p.bods)
	return itemsCopy
}

func (p *Product) GetActive() bool {
	return p.active
}

func (p *Product) CalculateRequirements(qty uint64) []valueObject.MaterialRequirement {
	var result []valueObject.MaterialRequirement

	for _, item := range p.bods {
		total := item.QuantityPerUnit * qty

		result = append(result, valueObject.MaterialRequirement{
			DrySupplyID: item.DrySupplyID,
			Quantity:    total,
		})
	}

	return result
}
