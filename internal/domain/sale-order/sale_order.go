package sale_order

import (
	valueObject "ricitelli-back/internal/value-object"
	"time"

	"github.com/google/uuid"
)

// SaleOrder represents the aggregate root that models a sales order made by a customer, including its status and associated items.
type SaleOrder struct {
	id         string
	customerID string
	status     string
	items      []valueObject.SaleOrderItem
	createdAt  string
	active     bool
}

func NewSaleOrder(customerID string, items []valueObject.SaleOrderItem) (SaleOrder, error) {
	return SaleOrder{
		id:         uuid.New().String(),
		customerID: customerID,
		status:     "NEW",
		items:      items,
		createdAt:  time.Now().UTC().Format(time.RFC3339),
		active:     true,
	}, nil
}

func (s *SaleOrder) GetID() string {
	return s.id
}

func (s *SaleOrder) GetCustomerID() string {
	return s.customerID
}

// GetItems ToDO: handle concurrency problems
func (s *SaleOrder) GetItems() []valueObject.SaleOrderItem {
	itemsCopy := make([]valueObject.SaleOrderItem, len(s.items))
	copy(itemsCopy, s.items)
	return itemsCopy
}

func (s *SaleOrder) GetCreatedAt() string {
	return s.createdAt
}
