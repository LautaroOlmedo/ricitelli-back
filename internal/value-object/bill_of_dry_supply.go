package value_object

// BillOfDrySupply represents the quantity of a specific dry supply required to produce one unit of a product.
type BillOfDrySupply struct {
	DrySupplyID string
	//DrySupplyName   string
	QuantityPerUnit uint64
}

func NewBillOfDrySupply(drySupplyID string, quantity uint64) BillOfDrySupply {
	if drySupplyID == "" {
		panic("drySupplyID cannot be empty")
	}
	if quantity <= 0 {
		panic("quantity must be greater than zero")
	}

	return BillOfDrySupply{
		DrySupplyID:     drySupplyID,
		QuantityPerUnit: quantity,
	}
}
