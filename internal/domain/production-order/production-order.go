package production_order

import (
	"errors"
	"ricitelli-back/internal/entities"
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusInProgress Status = "IN_PROGRESS"
	StatusCompleted  Status = "COMPLETED"
	StatusCancelled  Status = "CANCELLED"
)

var validTransitions = map[Status][]Status{
	StatusInProgress: {StatusCompleted, StatusCancelled},
}

type ProductionOrder struct {
	ID              string `json:"id"`
	salesOrderID    string
	operationNumber string
	status          Status
	items           []entities.ProductionItem
	createdAt       string
	active          bool
}

func NewProductionOrder(salesOrderID string, items []entities.ProductionItem) ProductionOrder {
	return ProductionOrder{
		ID:              uuid.New().String(),
		salesOrderID:    salesOrderID,
		operationNumber: "ProductionOrder",
		status:          StatusInProgress,
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

func (o *ProductionOrder) GetStatus() Status {
	return o.status
}

// UpdateStatus advances the production order through its lifecycle.
func (o *ProductionOrder) UpdateStatus(newStatus Status) error {
	allowed, ok := validTransitions[o.status]
	if !ok {
		return errors.New("production order is in a terminal state: " + string(o.status))
	}
	for _, a := range allowed {
		if a == newStatus {
			o.status = newStatus
			return nil
		}
	}
	return errors.New("invalid status transition from " + string(o.status) + " to " + string(newStatus))
}

// ReconstitueProductionOrder reconstitutes a ProductionOrder from stored data.
func ReconstitueProductionOrder(id, salesOrderID, operationNumber string, status Status, items []entities.ProductionItem, createdAt string, active bool) ProductionOrder {
	return ProductionOrder{
		ID:              id,
		salesOrderID:    salesOrderID,
		operationNumber: operationNumber,
		status:          status,
		items:           items,
		createdAt:       createdAt,
		active:          active,
	}
}
