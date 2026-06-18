package repository

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	product_inventory "ricitelli-back/internal/domain/product-inventory"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	valueObject "ricitelli-back/internal/value-object"
)

var _ sales_administration.Storage = (*InMemoryRepository)(nil)

func (r *InMemoryRepository) SaveSalesInvoice(_ context.Context, invoice sales_invoice.SalesInvoice) (*sales_invoice.SalesInvoice, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	if id, ok := store.salesInvoiceIdempotency[invoice.GetIdempotencyKey()]; ok {
		value := cloneSalesInvoice(store.salesInvoices[id])
		return &value, nil
	}
	customerID, err := r.saleOrderCustomerID(invoice.GetSaleOrderID())
	if err != nil {
		return nil, fmt.Errorf("sales invoice sale order: %w", err)
	}
	if customerID != invoice.GetCustomerID() {
		return nil, errors.New("sales invoice customer does not match sale order")
	}
	for _, current := range store.salesInvoices {
		if current.GetDocumentType() == invoice.GetDocumentType() &&
			current.GetPointOfSale() == invoice.GetPointOfSale() &&
			current.GetDocumentNumber() == invoice.GetDocumentNumber() {
			return nil, errors.New("sales invoice document already exists")
		}
	}
	value := cloneSalesInvoice(invoice)
	store.salesInvoices[value.GetID()] = value
	store.salesInvoiceIdempotency[value.GetIdempotencyKey()] = value.GetID()
	result := cloneSalesInvoice(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSalesInvoiceByID(_ context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.salesInvoices[id]
	if !ok {
		return nil, errors.New("sales invoice not found")
	}
	result := cloneSalesInvoice(value)
	return &result, nil
}

func (r *InMemoryRepository) ListSalesInvoices(_ context.Context, filter sales_administration.SalesInvoiceFilter) ([]sales_invoice.SalesInvoice, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]sales_invoice.SalesInvoice, 0)
	for _, invoice := range store.salesInvoices {
		if filter.CustomerID != "" && invoice.GetCustomerID() != filter.CustomerID ||
			filter.SaleOrderID != "" && invoice.GetSaleOrderID() != filter.SaleOrderID ||
			filter.Status != "" && string(invoice.GetStatus()) != filter.Status {
			continue
		}
		result = append(result, cloneSalesInvoice(invoice))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GetCreatedAt() > result[j].GetCreatedAt() })
	return result, nil
}

func (r *InMemoryRepository) IssueSalesInvoice(_ context.Context, id string) (*sales_invoice.SalesInvoice, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.salesInvoices[id]
	if !ok {
		return nil, errors.New("sales invoice not found")
	}
	if err := value.Issue(); err != nil {
		return nil, err
	}
	store.salesInvoices[id] = value
	result := cloneSalesInvoice(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSalesInvoiceOutstandingBalance(_ context.Context, id string) (string, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	return store.salesInvoiceOutstandingBalance(id)
}

func (r *InMemoryRepository) SaveRemittance(_ context.Context, document remittance.Remittance) (*remittance.Remittance, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	if id, ok := store.remittanceIdempotency[document.GetIdempotencyKey()]; ok {
		value := cloneRemittance(store.remittances[id])
		return &value, nil
	}
	customerID, err := r.saleOrderCustomerID(document.GetSaleOrderID())
	if err != nil {
		return nil, fmt.Errorf("remittance sale order: %w", err)
	}
	if customerID != document.GetCustomerID() {
		return nil, errors.New("remittance customer does not match sale order")
	}
	if document.GetSalesInvoiceID() != "" {
		invoice, ok := store.salesInvoices[document.GetSalesInvoiceID()]
		if !ok {
			return nil, errors.New("remittance sales invoice: sales invoice not found")
		}
		if invoice.GetCustomerID() != document.GetCustomerID() || invoice.GetSaleOrderID() != document.GetSaleOrderID() {
			return nil, errors.New("remittance invoice does not match customer and sale order")
		}
	}
	for _, current := range store.remittances {
		if current.GetPointOfSale() == document.GetPointOfSale() &&
			current.GetDocumentNumber() == document.GetDocumentNumber() {
			return nil, errors.New("remittance document already exists")
		}
	}
	value := cloneRemittance(document)
	store.remittances[value.GetID()] = value
	store.remittanceIdempotency[value.GetIdempotencyKey()] = value.GetID()
	result := cloneRemittance(value)
	return &result, nil
}

func (r *InMemoryRepository) GetRemittanceByID(_ context.Context, id string) (*remittance.Remittance, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.remittances[id]
	if !ok {
		return nil, errors.New("remittance not found")
	}
	result := cloneRemittance(value)
	return &result, nil
}

func (r *InMemoryRepository) ListRemittances(_ context.Context, filter sales_administration.RemittanceFilter) ([]remittance.Remittance, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]remittance.Remittance, 0)
	for _, document := range store.remittances {
		if filter.CustomerID != "" && document.GetCustomerID() != filter.CustomerID ||
			filter.SaleOrderID != "" && document.GetSaleOrderID() != filter.SaleOrderID ||
			filter.Status != "" && string(document.GetStatus()) != filter.Status {
			continue
		}
		result = append(result, cloneRemittance(document))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GetCreatedAt() > result[j].GetCreatedAt() })
	return result, nil
}

func (r *InMemoryRepository) ConfirmRemittance(_ context.Context, id, userID string) (*remittance.Remittance, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	document, ok := store.remittances[id]
	if !ok {
		return nil, errors.New("remittance not found")
	}
	if document.GetStatus() == remittance.StatusConfirmed {
		result := cloneRemittance(document)
		return &result, nil
	}

	r.productInventoryMutex.Lock()
	defer r.productInventoryMutex.Unlock()
	staged := make(map[int]product_inventory.ProductInventory)
	for _, item := range document.GetItems() {
		quantity, err := valueObject.DecimalToUint64(item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("remittance item quantity cannot be dispatched: %w", err)
		}
		index := -1
		for i := range r.ProductInventory {
			if r.ProductInventory[i].GetProductID() == item.ProductID {
				index = i
				break
			}
		}
		if index < 0 {
			return nil, errors.New("remittance product inventory: product inventory not found")
		}
		inventory, exists := staged[index]
		if !exists {
			inventory = cloneProductInventory(r.ProductInventory[index])
		}
		alreadyDispatched := false
		for _, movement := range inventory.GetMovements() {
			if movement.MovementType == valueObject.ProductDispatched && movement.Reference == id {
				alreadyDispatched = true
				break
			}
		}
		if !alreadyDispatched {
			if err := inventory.Dispatch(id, quantity, userID); err != nil {
				return nil, err
			}
		}
		staged[index] = inventory
	}
	if _, err := document.Confirm(); err != nil {
		return nil, err
	}
	if document.GetDeliveryDate() == "" {
		document = reconstituteRemittanceWithDeliveryDate(document, time.Now().UTC().Format(time.DateOnly))
	}
	for index, inventory := range staged {
		r.ProductInventory[index] = inventory
	}
	store.remittances[id] = document
	result := cloneRemittance(document)
	return &result, nil
}

func (r *InMemoryRepository) SaveCustomerReceipt(_ context.Context, receipt customer_receipt.CustomerReceipt) (*customer_receipt.CustomerReceipt, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()

	if id, ok := store.receiptIdempotency[receipt.GetIdempotencyKey()]; ok {
		value := cloneCustomerReceipt(store.customerReceipts[id])
		return &value, nil
	}
	for _, allocation := range receipt.GetAllocations() {
		invoice, ok := store.salesInvoices[allocation.SalesInvoiceID]
		if !ok {
			return nil, errors.New("receipt allocation invoice: sales invoice not found")
		}
		if invoice.GetCustomerID() != receipt.GetCustomerID() {
			return nil, errors.New("receipt allocation invoice belongs to another customer")
		}
		if invoice.GetCurrency() != receipt.GetCurrency() {
			return nil, errors.New("receipt allocation invoice uses another currency")
		}
		if invoice.GetStatus() == sales_invoice.StatusDraft || invoice.GetDocumentType() == sales_invoice.DocumentTypeCredit {
			return nil, errors.New("receipt allocations require an issued invoice or debit")
		}
	}
	for _, current := range store.customerReceipts {
		if current.GetReceiptNumber() == receipt.GetReceiptNumber() {
			return nil, errors.New("customer receipt number already exists")
		}
	}
	value := cloneCustomerReceipt(receipt)
	store.customerReceipts[value.GetID()] = value
	store.receiptIdempotency[value.GetIdempotencyKey()] = value.GetID()
	result := cloneCustomerReceipt(value)
	return &result, nil
}

func (r *InMemoryRepository) GetCustomerReceiptByID(_ context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.customerReceipts[id]
	if !ok {
		return nil, errors.New("customer receipt not found")
	}
	result := cloneCustomerReceipt(value)
	return &result, nil
}

func (r *InMemoryRepository) ListCustomerReceipts(_ context.Context, filter sales_administration.CustomerReceiptFilter) ([]customer_receipt.CustomerReceipt, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := make([]customer_receipt.CustomerReceipt, 0)
	for _, receipt := range store.customerReceipts {
		if filter.CustomerID != "" && receipt.GetCustomerID() != filter.CustomerID ||
			filter.Status != "" && string(receipt.GetStatus()) != filter.Status ||
			filter.SaleOrderID != "" && !receiptMatchesSaleOrder(store, receipt, filter.SaleOrderID) {
			continue
		}
		result = append(result, cloneCustomerReceipt(receipt))
	}
	sort.Slice(result, func(i, j int) bool { return result[i].GetCreatedAt() > result[j].GetCreatedAt() })
	return result, nil
}

func receiptMatchesSaleOrder(store *administrativeStore, receipt customer_receipt.CustomerReceipt, saleOrderID string) bool {
	for _, allocation := range receipt.GetAllocations() {
		if invoice, ok := store.salesInvoices[allocation.SalesInvoiceID]; ok && invoice.GetSaleOrderID() == saleOrderID {
			return true
		}
	}
	return false
}

func (r *InMemoryRepository) PostCustomerReceipt(_ context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	receipt, ok := store.customerReceipts[id]
	if !ok {
		return nil, errors.New("customer receipt not found")
	}
	if receipt.GetStatus() == customer_receipt.StatusPosted {
		result := cloneCustomerReceipt(receipt)
		return &result, nil
	}
	for _, allocation := range receipt.GetAllocations() {
		outstanding, err := store.salesInvoiceOutstandingBalance(allocation.SalesInvoiceID)
		if err != nil {
			return nil, err
		}
		if cmp, err := valueObject.CompareDecimal(allocation.AllocatedAmount, outstanding); err != nil || cmp > 0 {
			return nil, errors.New("receipt allocation exceeds invoice outstanding balance")
		}
	}
	if err := receipt.Post(); err != nil {
		return nil, err
	}
	store.customerReceipts[id] = receipt
	for _, allocation := range receipt.GetAllocations() {
		outstanding, err := store.salesInvoiceOutstandingBalance(allocation.SalesInvoiceID)
		if err != nil {
			return nil, err
		}
		invoice := store.salesInvoices[allocation.SalesInvoiceID]
		cmp, err := valueObject.CompareDecimal(outstanding, "0")
		if err != nil {
			return nil, err
		}
		invoice.UpdatePaymentStatus(cmp == 0)
		store.salesInvoices[allocation.SalesInvoiceID] = invoice
	}
	result := cloneCustomerReceipt(receipt)
	return &result, nil
}

func (s *administrativeStore) salesInvoiceOutstandingBalance(id string) (string, error) {
	invoice, ok := s.salesInvoices[id]
	if !ok {
		return "", errors.New("sales invoice not found")
	}
	if invoice.GetDocumentType() == sales_invoice.DocumentTypeCredit {
		return valueObject.SubtractDecimal("0", invoice.GetTotalAmount())
	}
	allocated := "0"
	for _, receipt := range s.customerReceipts {
		if receipt.GetStatus() != customer_receipt.StatusPosted {
			continue
		}
		for _, allocation := range receipt.GetAllocations() {
			if allocation.SalesInvoiceID == id {
				var err error
				allocated, err = valueObject.AddDecimal(allocated, allocation.AllocatedAmount)
				if err != nil {
					return "", err
				}
			}
		}
	}
	outstanding, err := valueObject.SubtractDecimal(invoice.GetTotalAmount(), allocated)
	if err != nil {
		return "", err
	}
	if cmp, _ := valueObject.CompareDecimal(outstanding, "0"); cmp < 0 {
		return "0", nil
	}
	return outstanding, nil
}

func (r *InMemoryRepository) saleOrderCustomerID(id string) (string, error) {
	r.saleOrdersMutex.RLock()
	defer r.saleOrdersMutex.RUnlock()
	for i := range r.SaleOrders {
		if r.SaleOrders[i].GetID() == id {
			return r.SaleOrders[i].GetCustomerID(), nil
		}
	}
	return "", errors.New("sale order not found")
}

func cloneSalesInvoice(value sales_invoice.SalesInvoice) sales_invoice.SalesInvoice {
	items := value.GetItems()
	sort.Slice(items, func(i, j int) bool { return items[i].LineNumber < items[j].LineNumber })
	return sales_invoice.ReconstituteSalesInvoice(value.GetID(), sales_invoice.NewSalesInvoiceParams{
		CustomerID: value.GetCustomerID(), SaleOrderID: value.GetSaleOrderID(), DocumentType: value.GetDocumentType(),
		PointOfSale: value.GetPointOfSale(), DocumentNumber: value.GetDocumentNumber(), IssueDate: value.GetIssueDate(),
		DueDate: value.GetDueDate(), Currency: value.GetCurrency(), ExchangeRate: value.GetExchangeRate(),
		Subtotal: value.GetSubtotal(), TaxTotal: value.GetTaxTotal(), TotalAmount: value.GetTotalAmount(),
		IdempotencyKey: value.GetIdempotencyKey(), Items: items,
	}, value.GetStatus(), value.GetCreatedAt(), value.GetUpdatedAt())
}

func cloneRemittance(value remittance.Remittance) remittance.Remittance {
	items := value.GetItems()
	sort.Slice(items, func(i, j int) bool { return items[i].LineNumber < items[j].LineNumber })
	return remittance.ReconstituteRemittance(value.GetID(), remittance.NewRemittanceParams{
		CustomerID: value.GetCustomerID(), SaleOrderID: value.GetSaleOrderID(), SalesInvoiceID: value.GetSalesInvoiceID(),
		PointOfSale: value.GetPointOfSale(), DocumentNumber: value.GetDocumentNumber(), IssueDate: value.GetIssueDate(),
		DeliveryDate: value.GetDeliveryDate(), IdempotencyKey: value.GetIdempotencyKey(), Items: items,
	}, value.GetStatus(), value.GetCreatedAt(), value.GetUpdatedAt())
}

func reconstituteRemittanceWithDeliveryDate(value remittance.Remittance, deliveryDate string) remittance.Remittance {
	return remittance.ReconstituteRemittance(value.GetID(), remittance.NewRemittanceParams{
		CustomerID: value.GetCustomerID(), SaleOrderID: value.GetSaleOrderID(), SalesInvoiceID: value.GetSalesInvoiceID(),
		PointOfSale: value.GetPointOfSale(), DocumentNumber: value.GetDocumentNumber(), IssueDate: value.GetIssueDate(),
		DeliveryDate: deliveryDate, IdempotencyKey: value.GetIdempotencyKey(), Items: value.GetItems(),
	}, value.GetStatus(), value.GetCreatedAt(), value.GetUpdatedAt())
}

func cloneCustomerReceipt(value customer_receipt.CustomerReceipt) customer_receipt.CustomerReceipt {
	allocations := value.GetAllocations()
	sort.Slice(allocations, func(i, j int) bool { return allocations[i].SalesInvoiceID < allocations[j].SalesInvoiceID })
	return customer_receipt.ReconstituteCustomerReceipt(value.GetID(), customer_receipt.NewCustomerReceiptParams{
		CustomerID: value.GetCustomerID(), ReceiptNumber: value.GetReceiptNumber(), ReceiptDate: value.GetReceiptDate(),
		Currency: value.GetCurrency(), ExchangeRate: value.GetExchangeRate(), Amount: value.GetAmount(),
		PaymentMethod: value.GetPaymentMethod(), PaymentReference: value.GetPaymentReference(),
		IdempotencyKey: value.GetIdempotencyKey(), Allocations: allocations,
	}, value.GetStatus(), value.GetCreatedAt(), value.GetUpdatedAt())
}

func cloneProductInventory(value product_inventory.ProductInventory) product_inventory.ProductInventory {
	return product_inventory.ReconstitueProductInventory(value.GetID(), value.GetProductID(), value.GetSku(), value.GetMovements())
}
