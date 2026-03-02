package production_order

import (
	"ricitelli-back/internal/entities"
	"time"

	"github.com/google/uuid"
)

type ProductionOrder struct {
	ID              string `json:"id"`
	salesOrderID    string
	operationNumber string
	status          string
	items           []entities.ProductionItem
	createdAt       string
	active          bool
}

/*type ProductSnapshot struct {
	productID   string
	productName string
	bomVersion  uint32
}*/

func NewProductionOrder(salesOrderID string, items []entities.ProductionItem) ProductionOrder {
	return ProductionOrder{
		ID:              uuid.New().String(),
		salesOrderID:    salesOrderID,
		operationNumber: "ProductionOrder",
		status:          "IN PROGRESS",
		items:           items,
		createdAt:       time.Now().String(), //.UTC().Format(time.RFC3339),
		active:          true,
	}

}

func (o *ProductionOrder) GetID() string {
	return o.ID
}

func (o *ProductionOrder) GetSalesOrderID() string {
	return o.salesOrderID
}

func (o *ProductionOrder) GetOperationNumber() string {
	return o.operationNumber
}

//func (o *ProductionOrder) GetDrySupplies() []*entities.DrySupply {
//	var drySupplies []*entities.DrySupply
//	for key, _ := range o.drySupplies {
//		drySupplies = append(drySupplies, key)
//	}
//	return drySupplies
//}
