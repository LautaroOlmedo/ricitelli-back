package value_object

import (
	"errors"
	"time"
)

type DrySupplyMovementType string

const (
	// DrySupplyIn: new stock arrived (purchase, inventory adjustment)
	DrySupplyIn DrySupplyMovementType = "DRY_SUPPLY_IN"
	// DrySupplyCommitted: reserved for a production order
	DrySupplyCommitted DrySupplyMovementType = "DRY_SUPPLY_COMMITTED"
	// DrySupplyReleased: reservation cancelled or production order cancelled
	DrySupplyReleased DrySupplyMovementType = "DRY_SUPPLY_RELEASED"
	// DrySupplyConsumed: actually used/consumed when production order completes
	DrySupplyConsumed DrySupplyMovementType = "DRY_SUPPLY_CONSUMED"
	// DrySupplyAdjusted: manual inventory correction
	DrySupplyAdjusted DrySupplyMovementType = "DRY_SUPPLY_ADJUSTED"
)

type DrySupplyMovement struct {
	MovementType DrySupplyMovementType
	Quantity     uint64
	Reference    string // production order ID or purchase reference
	CreatedAt    string
}

func NewDrySupplyMovement(movementType DrySupplyMovementType, quantity uint64, reference string) (DrySupplyMovement, error) {
	if movementType == "" {
		return DrySupplyMovement{}, errors.New("movement type cannot be empty")
	}
	if quantity == 0 {
		return DrySupplyMovement{}, errors.New("quantity cannot be 0")
	}
	return DrySupplyMovement{
		MovementType: movementType,
		Quantity:     quantity,
		Reference:    reference,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}
