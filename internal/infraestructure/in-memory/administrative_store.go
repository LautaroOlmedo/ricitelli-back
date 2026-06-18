package repository

import (
	"sync"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	purchasing "ricitelli-back/internal/domain/purchasing-administration"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
)

// administrativeStore keeps the newly added administrative slices isolated
// from the legacy seeded repository data. A single mutex provides transaction-
// like atomicity for posting, allocation, balance, and VAT operations.
type administrativeStore struct {
	mu sync.RWMutex

	salesInvoices           map[string]sales_invoice.SalesInvoice
	salesInvoiceIdempotency map[string]string
	remittances             map[string]remittance.Remittance
	remittanceIdempotency   map[string]string
	customerReceipts        map[string]customer_receipt.CustomerReceipt
	receiptIdempotency      map[string]string

	suppliers                  map[string]purchasing.Supplier
	supplierIdempotency        map[string]string
	purchaseNeeds              map[string]purchasing.PurchaseNeed
	purchaseNeedIdempotency    map[string]string
	supplierQuotes             map[string]purchasing.SupplierQuote
	supplierQuoteIdempotency   map[string]string
	supplierInvoices           map[string]purchasing.SupplierInvoice
	supplierInvoiceIdempotency map[string]string
	supplierPayments           map[string]purchasing.SupplierPayment
	supplierPaymentIdempotency map[string]string
}

func newAdministrativeStore() *administrativeStore {
	return &administrativeStore{
		salesInvoices:              make(map[string]sales_invoice.SalesInvoice),
		salesInvoiceIdempotency:    make(map[string]string),
		remittances:                make(map[string]remittance.Remittance),
		remittanceIdempotency:      make(map[string]string),
		customerReceipts:           make(map[string]customer_receipt.CustomerReceipt),
		receiptIdempotency:         make(map[string]string),
		suppliers:                  make(map[string]purchasing.Supplier),
		supplierIdempotency:        make(map[string]string),
		purchaseNeeds:              make(map[string]purchasing.PurchaseNeed),
		purchaseNeedIdempotency:    make(map[string]string),
		supplierQuotes:             make(map[string]purchasing.SupplierQuote),
		supplierQuoteIdempotency:   make(map[string]string),
		supplierInvoices:           make(map[string]purchasing.SupplierInvoice),
		supplierInvoiceIdempotency: make(map[string]string),
		supplierPayments:           make(map[string]purchasing.SupplierPayment),
		supplierPaymentIdempotency: make(map[string]string),
	}
}

func (r *InMemoryRepository) adminStore() *administrativeStore {
	r.administrationOnce.Do(func() {
		r.administration = newAdministrativeStore()
	})
	return r.administration
}
