package value_object

// SaleOrderItem represents an item within a sales order, indicating the product sold, the quantity and its unit price.
type SaleOrderItem struct {
	ProductID string
	Quantity  uint64
	UnitPrice float32 // BigDecimal?
}
