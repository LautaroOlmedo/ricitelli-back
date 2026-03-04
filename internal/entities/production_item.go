package entities

import valueObject "ricitelli-back/internal/value-object"

type ProductionItem struct {
	// ID string
	ProductID    string
	Quantity     uint64
	Requirements []valueObject.MaterialRequirement
}
