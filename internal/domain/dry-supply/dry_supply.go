package dry_supply

import (
	"errors"

	"github.com/google/uuid"
)

type Category string

const (
	CategoryLabel          Category = "LABEL"
	CategoryContraetiqueta Category = "CONTRAETIQUETA"
	CategoryBox            Category = "BOX"
	CategoryCork           Category = "CORK"
	CategoryCapsule        Category = "CAPSULE"
	CategoryBottle         Category = "BOTTLE"
	CategoryOther          Category = "OTHER"
)

// DrySupply represents a raw/packaging material (insumo seco).
// Examples: "IF1156" caja Hey Malbec!, etiqueta Japón, cápsula Kung Fu.
type DrySupply struct {
	id          string
	code        string   // SKU/code, e.g., "IF1156"
	name        string   // human-readable name
	category    Category // LABEL, BOX, CORK, etc.
	description string
	unit        string // "UNIT", "BOX", "KG"
}

func NewDrySupply(code, name string, category Category, unit string) (DrySupply, error) {
	if code == "" {
		return DrySupply{}, errors.New("dry supply code cannot be empty")
	}
	if name == "" {
		return DrySupply{}, errors.New("dry supply name cannot be empty")
	}
	return DrySupply{
		id:       uuid.New().String(),
		code:     code,
		name:     name,
		category: category,
		unit:     unit,
	}, nil
}

func NewDrySupplyWithID(id, code, name string, category Category, unit string) (DrySupply, error) {
	if id == "" {
		return DrySupply{}, errors.New("id cannot be empty")
	}
	ds, err := NewDrySupply(code, name, category, unit)
	if err != nil {
		return DrySupply{}, err
	}
	ds.id = id
	return ds, nil
}

func (d *DrySupply) GetID() string          { return d.id }
func (d *DrySupply) GetCode() string        { return d.code }
func (d *DrySupply) GetName() string        { return d.name }
func (d *DrySupply) GetCategory() Category  { return d.category }
func (d *DrySupply) GetDescription() string { return d.description }
func (d *DrySupply) GetUnit() string        { return d.unit }
