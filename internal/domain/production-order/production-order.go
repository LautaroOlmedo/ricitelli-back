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

func NewProductionOrder(salesOrderID string, items []entities.ProductionItem) ProductionOrder {
	return ProductionOrder{
		ID:              uuid.New().String(),
		salesOrderID:    salesOrderID,
		operationNumber: "ProductionOrder",
		status:          "IN PROGRESS",
		items:           items,
		createdAt:       time.Now().UTC().Format(time.RFC3339),
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

func (o *ProductionOrder) GetCreatedAt() string {
	return o.createdAt
}

func (o *ProductionOrder) GetItems() []entities.ProductionItem {
	itemsCopy := make([]entities.ProductionItem, len(o.items))
	copy(itemsCopy, o.items)
	return itemsCopy
}

func (o *ProductionOrder) GetStatus() string {
	return o.status
}

//func (o *ProductionOrder) GetDrySupplies() []*entities.DrySupply {
//	var drySupplies []*entities.DrySupply
//	for key, _ := range o.drySupplies {
//		drySupplies = append(drySupplies, key)
//	}
//	return drySupplies
//}

/*type ProductSnapshot struct {
	productID   string
	productName string
	bomVersion  uint32
}*/
