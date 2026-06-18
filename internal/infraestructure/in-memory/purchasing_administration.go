package repository

import (
	"context"
	"errors"
	"sort"
	"time"

	purchasing "ricitelli-back/internal/domain/purchasing-administration"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	purchasing_administration "ricitelli-back/internal/service/purchasing-administration"
)

var _ purchasing_administration.Storage = (*InMemoryRepository)(nil)

func (r *InMemoryRepository) CreateSupplier(_ context.Context, params purchasing.NewSupplierParams) (purchasing.Supplier, error) {
	value, err := purchasing.NewSupplier(params)
	if err != nil {
		return purchasing.Supplier{}, err
	}
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, ok := store.supplierIdempotency[value.IdempotencyKey]; ok {
		return store.suppliers[id], nil
	}
	for _, current := range store.suppliers {
		if current.TaxID == value.TaxID {
			return purchasing.Supplier{}, errors.New("supplier tax_id already exists")
		}
	}
	store.suppliers[value.ID] = value
	store.supplierIdempotency[value.IdempotencyKey] = value.ID
	return value, nil
}

func (r *InMemoryRepository) GetSupplierByID(_ context.Context, id string) (*purchasing.Supplier, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.suppliers[id]
	if !ok {
		return nil, errors.New("supplier not found")
	}
	return &value, nil
}

func (r *InMemoryRepository) GetSuppliers(_ context.Context, includeInactive bool) ([]purchasing.Supplier, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []purchasing.Supplier
	for _, value := range store.suppliers {
		if includeInactive || value.Active {
			result = append(result, value)
		}
	}
	sort.Slice(result, func(i, j int) bool { return result[i].SocialReason < result[j].SocialReason })
	return result, nil
}

func (r *InMemoryRepository) UpdateSupplier(_ context.Context, id string, params purchasing.UpdateSupplierParams) (*purchasing.Supplier, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.suppliers[id]
	if !ok {
		return nil, errors.New("supplier not found")
	}
	if params.TaxID != "" && params.TaxID != value.TaxID {
		for otherID, current := range store.suppliers {
			if otherID != id && current.TaxID == params.TaxID {
				return nil, errors.New("supplier tax_id already exists")
			}
		}
	}
	value.Update(params)
	store.suppliers[id] = value
	return &value, nil
}

func (r *InMemoryRepository) DeactivateSupplier(_ context.Context, id string) (*purchasing.Supplier, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.suppliers[id]
	if !ok {
		return nil, errors.New("supplier not found")
	}
	value.Active = false
	value.UpdatedAt = administrativeNow()
	store.suppliers[id] = value
	return &value, nil
}

func (r *InMemoryRepository) CreatePurchaseNeed(_ context.Context, params purchasing.NewPurchaseNeedParams) (purchasing.PurchaseNeed, error) {
	value, err := purchasing.NewPurchaseNeed(params)
	if err != nil {
		return purchasing.PurchaseNeed{}, err
	}
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, ok := store.purchaseNeedIdempotency[value.IdempotencyKey]; ok {
		return clonePurchaseNeed(store.purchaseNeeds[id]), nil
	}
	for _, current := range store.purchaseNeeds {
		if current.NeedNumber == value.NeedNumber {
			return purchasing.PurchaseNeed{}, errors.New("purchase need number already exists")
		}
	}
	value = clonePurchaseNeed(value)
	store.purchaseNeeds[value.ID] = value
	store.purchaseNeedIdempotency[value.IdempotencyKey] = value.ID
	return clonePurchaseNeed(value), nil
}

func (r *InMemoryRepository) GetPurchaseNeedByID(_ context.Context, id string) (*purchasing.PurchaseNeed, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.purchaseNeeds[id]
	if !ok {
		return nil, errors.New("purchase need not found")
	}
	result := clonePurchaseNeed(value)
	return &result, nil
}

func (r *InMemoryRepository) GetPurchaseNeeds(_ context.Context) ([]purchasing.PurchaseNeed, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []purchasing.PurchaseNeed
	for _, value := range store.purchaseNeeds {
		result = append(result, clonePurchaseNeed(value))
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].RequestedDate != result[j].RequestedDate {
			return result[i].RequestedDate > result[j].RequestedDate
		}
		return result[i].NeedNumber > result[j].NeedNumber
	})
	return result, nil
}

func (r *InMemoryRepository) UpdatePurchaseNeedStatus(_ context.Context, id string, status purchasing.PurchaseNeedStatus) (*purchasing.PurchaseNeed, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.purchaseNeeds[id]
	if !ok {
		return nil, errors.New("purchase need not found")
	}
	value.Status, value.UpdatedAt = status, administrativeNow()
	store.purchaseNeeds[id] = value
	result := clonePurchaseNeed(value)
	return &result, nil
}

func (r *InMemoryRepository) CreateSupplierQuote(_ context.Context, params purchasing.NewSupplierQuoteParams) (purchasing.SupplierQuote, error) {
	value, err := purchasing.NewSupplierQuote(params)
	if err != nil {
		return purchasing.SupplierQuote{}, err
	}
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, ok := store.supplierQuoteIdempotency[value.IdempotencyKey]; ok {
		return cloneSupplierQuote(store.supplierQuotes[id]), nil
	}
	if _, ok := store.suppliers[value.SupplierID]; !ok {
		return purchasing.SupplierQuote{}, errors.New("supplier not found")
	}
	if value.PurchaseNeedID != "" {
		if _, ok := store.purchaseNeeds[value.PurchaseNeedID]; !ok {
			return purchasing.SupplierQuote{}, errors.New("purchase need not found")
		}
	}
	for _, current := range store.supplierQuotes {
		if current.SupplierID == value.SupplierID && current.QuoteNumber == value.QuoteNumber {
			return purchasing.SupplierQuote{}, errors.New("supplier quote number already exists")
		}
	}
	value = cloneSupplierQuote(value)
	store.supplierQuotes[value.ID] = value
	store.supplierQuoteIdempotency[value.IdempotencyKey] = value.ID
	return cloneSupplierQuote(value), nil
}

func (r *InMemoryRepository) GetSupplierQuoteByID(_ context.Context, id string) (*purchasing.SupplierQuote, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.supplierQuotes[id]
	if !ok {
		return nil, errors.New("supplier quote not found")
	}
	result := cloneSupplierQuote(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSupplierQuotes(_ context.Context, supplierID string) ([]purchasing.SupplierQuote, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []purchasing.SupplierQuote
	for _, value := range store.supplierQuotes {
		if supplierID == "" || value.SupplierID == supplierID {
			result = append(result, cloneSupplierQuote(value))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].QuoteDate != result[j].QuoteDate {
			return result[i].QuoteDate > result[j].QuoteDate
		}
		return result[i].CreatedAt > result[j].CreatedAt
	})
	return result, nil
}

func (r *InMemoryRepository) UpdateSupplierQuoteStatus(_ context.Context, id string, status purchasing.SupplierQuoteStatus) (*purchasing.SupplierQuote, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.supplierQuotes[id]
	if !ok {
		return nil, errors.New("supplier quote not found")
	}
	value.Status, value.UpdatedAt = status, administrativeNow()
	store.supplierQuotes[id] = value
	result := cloneSupplierQuote(value)
	return &result, nil
}

func (r *InMemoryRepository) CreateSupplierInvoice(_ context.Context, params purchasing.NewSupplierInvoiceParams) (purchasing.SupplierInvoice, error) {
	value, err := purchasing.NewSupplierInvoice(params)
	if err != nil {
		return purchasing.SupplierInvoice{}, err
	}
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, ok := store.supplierInvoiceIdempotency[value.IdempotencyKey]; ok {
		return cloneSupplierInvoice(store.supplierInvoices[id]), nil
	}
	if _, ok := store.suppliers[value.SupplierID]; !ok {
		return purchasing.SupplierInvoice{}, errors.New("supplier not found")
	}
	if value.SupplierQuoteID != "" {
		quote, ok := store.supplierQuotes[value.SupplierQuoteID]
		if !ok {
			return purchasing.SupplierInvoice{}, errors.New("supplier quote not found")
		}
		if quote.SupplierID != value.SupplierID {
			return purchasing.SupplierInvoice{}, errors.New("supplier invoice quote belongs to another supplier")
		}
	}
	for _, current := range store.supplierInvoices {
		if current.SupplierID == value.SupplierID && current.DocumentType == value.DocumentType &&
			current.PointOfSale == value.PointOfSale && current.DocumentNumber == value.DocumentNumber {
			return purchasing.SupplierInvoice{}, errors.New("supplier invoice document already exists")
		}
	}
	value = cloneSupplierInvoice(value)
	store.supplierInvoices[value.ID] = value
	store.supplierInvoiceIdempotency[value.IdempotencyKey] = value.ID
	return cloneSupplierInvoice(value), nil
}

func (r *InMemoryRepository) GetSupplierInvoiceByID(_ context.Context, id string) (*purchasing.SupplierInvoice, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.supplierInvoices[id]
	if !ok {
		return nil, errors.New("supplier invoice not found")
	}
	result := cloneSupplierInvoice(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSupplierInvoices(_ context.Context, supplierID string) ([]purchasing.SupplierInvoice, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []purchasing.SupplierInvoice
	for _, value := range store.supplierInvoices {
		if supplierID == "" || value.SupplierID == supplierID {
			result = append(result, cloneSupplierInvoice(value))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].IssueDate != result[j].IssueDate {
			return result[i].IssueDate > result[j].IssueDate
		}
		return result[i].CreatedAt > result[j].CreatedAt
	})
	return result, nil
}

func (r *InMemoryRepository) UpdateSupplierInvoiceStatus(_ context.Context, id string, status purchasing.SupplierInvoiceStatus) (*purchasing.SupplierInvoice, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.supplierInvoices[id]
	if !ok {
		return nil, errors.New("supplier invoice not found")
	}
	value.Status, value.UpdatedAt = status, administrativeNow()
	store.supplierInvoices[id] = value
	result := cloneSupplierInvoice(value)
	return &result, nil
}

func (r *InMemoryRepository) CreateSupplierPayment(_ context.Context, params purchasing.NewSupplierPaymentParams) (purchasing.SupplierPayment, error) {
	value, err := purchasing.NewSupplierPayment(params)
	if err != nil {
		return purchasing.SupplierPayment{}, err
	}
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	if id, ok := store.supplierPaymentIdempotency[value.IdempotencyKey]; ok {
		return cloneSupplierPayment(store.supplierPayments[id]), nil
	}
	if _, ok := store.suppliers[value.SupplierID]; !ok {
		return purchasing.SupplierPayment{}, errors.New("supplier not found")
	}
	for _, allocation := range value.Allocations {
		invoice, ok := store.supplierInvoices[allocation.SupplierInvoiceID]
		if !ok {
			return purchasing.SupplierPayment{}, errors.New("supplier invoice not found")
		}
		if invoice.SupplierID != value.SupplierID {
			return purchasing.SupplierPayment{}, errors.New("supplier payment invoice belongs to another supplier")
		}
	}
	for _, current := range store.supplierPayments {
		if current.PaymentNumber == value.PaymentNumber {
			return purchasing.SupplierPayment{}, errors.New("supplier payment number already exists")
		}
	}
	value = cloneSupplierPayment(value)
	store.supplierPayments[value.ID] = value
	store.supplierPaymentIdempotency[value.IdempotencyKey] = value.ID
	return cloneSupplierPayment(value), nil
}

func (r *InMemoryRepository) GetSupplierPaymentByID(_ context.Context, id string) (*purchasing.SupplierPayment, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	value, ok := store.supplierPayments[id]
	if !ok {
		return nil, errors.New("supplier payment not found")
	}
	result := cloneSupplierPayment(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSupplierPayments(_ context.Context, supplierID string) ([]purchasing.SupplierPayment, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	var result []purchasing.SupplierPayment
	for _, value := range store.supplierPayments {
		if supplierID == "" || value.SupplierID == supplierID {
			result = append(result, cloneSupplierPayment(value))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].PaymentDate != result[j].PaymentDate {
			return result[i].PaymentDate > result[j].PaymentDate
		}
		return result[i].CreatedAt > result[j].CreatedAt
	})
	return result, nil
}

func (r *InMemoryRepository) UpdateSupplierPaymentStatus(_ context.Context, id string, status purchasing.SupplierPaymentStatus) (*purchasing.SupplierPayment, error) {
	store := r.adminStore()
	store.mu.Lock()
	defer store.mu.Unlock()
	value, ok := store.supplierPayments[id]
	if !ok {
		return nil, errors.New("supplier payment not found")
	}
	now := administrativeNow()
	value.Status, value.UpdatedAt = status, now
	store.supplierPayments[id] = value
	for _, allocation := range value.Allocations {
		invoice, ok := store.supplierInvoices[allocation.SupplierInvoiceID]
		if !ok || invoice.Status == purchasing.SupplierInvoiceDraft || invoice.Status == purchasing.SupplierInvoiceCancelled {
			continue
		}
		allocated, err := store.postedSupplierInvoiceAllocation(invoice.ID)
		if err != nil {
			return nil, err
		}
		cmp, err := purchasing.CompareDecimals(allocated, invoice.TotalAmount)
		if err != nil {
			return nil, err
		}
		switch {
		case cmp >= 0:
			invoice.Status = purchasing.SupplierInvoicePaid
		case allocated != "0":
			invoice.Status = purchasing.SupplierInvoicePartiallyPaid
		default:
			invoice.Status = purchasing.SupplierInvoiceIssued
		}
		invoice.UpdatedAt = now
		store.supplierInvoices[invoice.ID] = invoice
	}
	result := cloneSupplierPayment(value)
	return &result, nil
}

func (r *InMemoryRepository) GetSupplierOutstandingBalance(_ context.Context, supplierID string, currency purchasing.Currency) (*purchasing.SupplierOutstandingBalance, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	result := &purchasing.SupplierOutstandingBalance{SupplierID: supplierID, Currency: string(currency), TotalOutstanding: "0"}
	for _, invoice := range store.supplierInvoices {
		if invoice.SupplierID != supplierID || invoice.Currency != currency ||
			(invoice.Status != purchasing.SupplierInvoiceIssued &&
				invoice.Status != purchasing.SupplierInvoicePartiallyPaid &&
				invoice.Status != purchasing.SupplierInvoicePaid) {
			continue
		}
		total := invoice.TotalAmount
		if invoice.DocumentType == purchasing.SupplierInvoiceTypeCreditNote {
			var err error
			total, err = purchasing.SubtractDecimals("0", total)
			if err != nil {
				return nil, err
			}
		}
		allocated, err := store.postedSupplierInvoiceAllocation(invoice.ID)
		if err != nil {
			return nil, err
		}
		outstanding, err := purchasing.SubtractDecimals(total, allocated)
		if err != nil {
			return nil, err
		}
		result.Invoices = append(result.Invoices, purchasing.SupplierInvoiceBalance{
			SupplierInvoiceID: invoice.ID, SupplierID: invoice.SupplierID, DocumentType: string(invoice.DocumentType),
			PointOfSale: invoice.PointOfSale, DocumentNumber: invoice.DocumentNumber, IssueDate: invoice.IssueDate,
			Currency: string(invoice.Currency), TotalAmount: total, AllocatedAmount: allocated, OutstandingAmount: outstanding,
		})
		result.TotalOutstanding, err = purchasing.AddDecimals(result.TotalOutstanding, outstanding)
		if err != nil {
			return nil, err
		}
	}
	sort.Slice(result.Invoices, func(i, j int) bool {
		if result.Invoices[i].IssueDate != result.Invoices[j].IssueDate {
			return result.Invoices[i].IssueDate < result.Invoices[j].IssueDate
		}
		return result.Invoices[i].DocumentNumber < result.Invoices[j].DocumentNumber
	})
	return result, nil
}

func (r *InMemoryRepository) GetMonthlyVATPosition(_ context.Context, month string) (*purchasing.MonthlyVATPosition, error) {
	store := r.adminStore()
	store.mu.RLock()
	defer store.mu.RUnlock()
	type totals struct{ sales, purchases string }
	byRate := make(map[string]totals)
	for _, invoice := range store.salesInvoices {
		if invoice.GetStatus() == sales_invoice.StatusDraft || len(invoice.GetIssueDate()) < 7 || invoice.GetIssueDate()[:7] != month {
			continue
		}
		for _, item := range invoice.GetItems() {
			rate, err := purchasing.NormalizeDecimal(item.TaxRate)
			if err != nil {
				return nil, err
			}
			current := byRate[rate]
			if current.sales == "" {
				current.sales = "0"
			}
			if current.purchases == "" {
				current.purchases = "0"
			}
			if invoice.GetDocumentType() == sales_invoice.DocumentTypeCredit {
				current.sales, err = purchasing.SubtractDecimals(current.sales, item.TaxAmount)
			} else {
				current.sales, err = purchasing.AddDecimals(current.sales, item.TaxAmount)
			}
			if err != nil {
				return nil, err
			}
			byRate[rate] = current
		}
	}
	for _, invoice := range store.supplierInvoices {
		if invoice.Status == purchasing.SupplierInvoiceDraft || invoice.Status == purchasing.SupplierInvoiceCancelled ||
			len(invoice.IssueDate) < 7 || invoice.IssueDate[:7] != month {
			continue
		}
		for _, item := range invoice.Items {
			rate, err := purchasing.NormalizeDecimal(item.TaxRate)
			if err != nil {
				return nil, err
			}
			current := byRate[rate]
			if current.sales == "" {
				current.sales = "0"
			}
			if current.purchases == "" {
				current.purchases = "0"
			}
			if invoice.DocumentType == purchasing.SupplierInvoiceTypeCreditNote {
				current.purchases, err = purchasing.SubtractDecimals(current.purchases, item.TaxAmount)
			} else {
				current.purchases, err = purchasing.AddDecimals(current.purchases, item.TaxAmount)
			}
			if err != nil {
				return nil, err
			}
			byRate[rate] = current
		}
	}
	lines := make([]purchasing.VATPositionLine, 0, len(byRate))
	for rate, value := range byRate {
		lines = append(lines, purchasing.VATPositionLine{TaxRate: rate, SalesDebit: value.sales, PurchaseCredit: value.purchases})
	}
	sort.Slice(lines, func(i, j int) bool {
		cmp, _ := purchasing.CompareDecimals(lines[i].TaxRate, lines[j].TaxRate)
		return cmp < 0
	})
	result, err := purchasing.NewMonthlyVATPosition(month, lines)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *administrativeStore) postedSupplierInvoiceAllocation(invoiceID string) (string, error) {
	total := "0"
	for _, payment := range s.supplierPayments {
		if payment.Status != purchasing.SupplierPaymentPosted {
			continue
		}
		for _, allocation := range payment.Allocations {
			if allocation.SupplierInvoiceID == invoiceID {
				var err error
				total, err = purchasing.AddDecimals(total, allocation.AllocatedAmount)
				if err != nil {
					return "", err
				}
			}
		}
	}
	return total, nil
}

func administrativeNow() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}

func clonePurchaseNeed(value purchasing.PurchaseNeed) purchasing.PurchaseNeed {
	value.Items = append([]purchasing.PurchaseNeedItem(nil), value.Items...)
	sort.Slice(value.Items, func(i, j int) bool { return value.Items[i].LineNumber < value.Items[j].LineNumber })
	return value
}

func cloneSupplierQuote(value purchasing.SupplierQuote) purchasing.SupplierQuote {
	value.Items = append([]purchasing.SupplierQuoteItem(nil), value.Items...)
	sort.Slice(value.Items, func(i, j int) bool { return value.Items[i].LineNumber < value.Items[j].LineNumber })
	return value
}

func cloneSupplierInvoice(value purchasing.SupplierInvoice) purchasing.SupplierInvoice {
	value.Items = append([]purchasing.SupplierInvoiceItem(nil), value.Items...)
	sort.Slice(value.Items, func(i, j int) bool { return value.Items[i].LineNumber < value.Items[j].LineNumber })
	return value
}

func cloneSupplierPayment(value purchasing.SupplierPayment) purchasing.SupplierPayment {
	value.Allocations = append([]purchasing.SupplierPaymentAllocation(nil), value.Allocations...)
	sort.Slice(value.Allocations, func(i, j int) bool {
		if value.Allocations[i].CreatedAt != value.Allocations[j].CreatedAt {
			return value.Allocations[i].CreatedAt < value.Allocations[j].CreatedAt
		}
		return value.Allocations[i].ID < value.Allocations[j].ID
	})
	return value
}
