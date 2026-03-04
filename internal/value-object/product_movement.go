package value_object

import (
	"errors"
	"time"
)

type ProductMovementType string
type Stage string

const (
	// Producción
	//PRODUCT_PRODUCED

	//Cuando termina producción inicial. Ej: +100 SIN_VESTIR
	ProductProduced ProductMovementType = "PRODUCT_PRODUCED"

	// Transformación de estado. PRODUCT_STAGE_OUT / PRODUCT_STAGE_IN Cuando se viste: -100 SIN_VESTIR (STAGE_OUT) +100 VESTIDO (STAGE_IN)
	ProductStageIn  ProductMovementType = "PRODUCT_STAGE_IN"
	ProductStageOut ProductMovementType = "PRODUCT_STAGE_OUT"

	// Venta PRODUCT_RESERVED_FOR_SALE. cuando se crea SaleOrder: -30 VESTIDO (RESERVED). No cambia stock físico. Solo compromete disponibilidad.
	// PRODUCT_RESERVATION_RELEASED. Si se cancela la orden: +30 VESTIDO
	ProductReservedForSale     ProductMovementType = "PRODUCT_RESERVED_FOR_SALE"
	ProductReservationReleased ProductMovementType = "PRODUCT_RESERVATION_RELEASED"

	// Despacho
	ProductDispatched ProductMovementType = "PRODUCT_DISPATCHED"

	// Ajustes
	ProductStockAdjusted ProductMovementType = "PRODUCT_STOCK_ADJUSTED"

	Dressed   Stage = "DRESSED"
	Undressed Stage = "UNDRESSED"
)

type ProductMovement struct {
	MovementType ProductMovementType
	Quantity     uint64
	Reference    string // saleOrderID, productionOrderID, dispatchID
	Stage        Stage  // VESTIDO, SIN VESTIR
	CreatedAt    string
}

func NewProductMovement(reference string, stage Stage, movementType ProductMovementType, quantity uint64) (ProductMovement, error) {
	if movementType == "" || reference == "" {
		return ProductMovement{}, errors.New("inventory movement type or reference is empty")
	}
	if quantity == 0 {
		return ProductMovement{}, errors.New("quantity cannot be 0")
	}
	return ProductMovement{
		Reference:    reference,
		MovementType: movementType,
		Stage:        stage,
		Quantity:     quantity,
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}
