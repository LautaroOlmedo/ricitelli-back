package purchasing_administration

import (
	"context"

	domain "ricitelli-back/internal/domain/purchasing-administration"
)

// Storage is the persistence boundary for the complete purchasing-administration slice.
type Storage interface {
	CreateSupplier(context.Context, domain.NewSupplierParams) (domain.Supplier, error)
	GetSupplierByID(context.Context, string) (*domain.Supplier, error)
	GetSuppliers(context.Context, bool) ([]domain.Supplier, error)
	UpdateSupplier(context.Context, string, domain.UpdateSupplierParams) (*domain.Supplier, error)
	DeactivateSupplier(context.Context, string) (*domain.Supplier, error)

	CreatePurchaseNeed(context.Context, domain.NewPurchaseNeedParams) (domain.PurchaseNeed, error)
	GetPurchaseNeedByID(context.Context, string) (*domain.PurchaseNeed, error)
	GetPurchaseNeeds(context.Context) ([]domain.PurchaseNeed, error)
	UpdatePurchaseNeedStatus(context.Context, string, domain.PurchaseNeedStatus) (*domain.PurchaseNeed, error)

	CreateSupplierQuote(context.Context, domain.NewSupplierQuoteParams) (domain.SupplierQuote, error)
	GetSupplierQuoteByID(context.Context, string) (*domain.SupplierQuote, error)
	GetSupplierQuotes(context.Context, string) ([]domain.SupplierQuote, error)
	UpdateSupplierQuoteStatus(context.Context, string, domain.SupplierQuoteStatus) (*domain.SupplierQuote, error)

	CreateSupplierInvoice(context.Context, domain.NewSupplierInvoiceParams) (domain.SupplierInvoice, error)
	GetSupplierInvoiceByID(context.Context, string) (*domain.SupplierInvoice, error)
	GetSupplierInvoices(context.Context, string) ([]domain.SupplierInvoice, error)
	UpdateSupplierInvoiceStatus(context.Context, string, domain.SupplierInvoiceStatus) (*domain.SupplierInvoice, error)

	CreateSupplierPayment(context.Context, domain.NewSupplierPaymentParams) (domain.SupplierPayment, error)
	GetSupplierPaymentByID(context.Context, string) (*domain.SupplierPayment, error)
	GetSupplierPayments(context.Context, string) ([]domain.SupplierPayment, error)
	UpdateSupplierPaymentStatus(context.Context, string, domain.SupplierPaymentStatus) (*domain.SupplierPayment, error)
	GetSupplierOutstandingBalance(context.Context, string, domain.Currency) (*domain.SupplierOutstandingBalance, error)

	GetMonthlyVATPosition(context.Context, string) (*domain.MonthlyVATPosition, error)
}

type Service struct{ Storage Storage }

func NewService(storage Storage) *Service { return &Service{Storage: storage} }
