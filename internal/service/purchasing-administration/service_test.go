package purchasing_administration

import (
	"context"
	"errors"
	"testing"

	domain "ricitelli-back/internal/domain/purchasing-administration"

	"github.com/stretchr/testify/require"
)

type testStorage struct {
	invoice              domain.SupplierInvoice
	payment              domain.SupplierPayment
	updatedInvoiceStatus domain.SupplierInvoiceStatus
	updatedPaymentStatus domain.SupplierPaymentStatus
	vatCalls             int
}

func (s *testStorage) CreateSupplier(context.Context, domain.NewSupplierParams) (domain.Supplier, error) {
	return domain.Supplier{}, nil
}
func (s *testStorage) GetSupplierByID(context.Context, string) (*domain.Supplier, error) {
	return nil, errors.New("unused")
}
func (s *testStorage) GetSuppliers(context.Context, bool) ([]domain.Supplier, error) { return nil, nil }
func (s *testStorage) UpdateSupplier(context.Context, string, domain.UpdateSupplierParams) (*domain.Supplier, error) {
	return nil, nil
}
func (s *testStorage) DeactivateSupplier(context.Context, string) (*domain.Supplier, error) {
	return nil, nil
}
func (s *testStorage) CreatePurchaseNeed(context.Context, domain.NewPurchaseNeedParams) (domain.PurchaseNeed, error) {
	return domain.PurchaseNeed{}, nil
}
func (s *testStorage) GetPurchaseNeedByID(context.Context, string) (*domain.PurchaseNeed, error) {
	return nil, errors.New("unused")
}
func (s *testStorage) GetPurchaseNeeds(context.Context) ([]domain.PurchaseNeed, error) {
	return nil, nil
}
func (s *testStorage) UpdatePurchaseNeedStatus(context.Context, string, domain.PurchaseNeedStatus) (*domain.PurchaseNeed, error) {
	return nil, nil
}
func (s *testStorage) CreateSupplierQuote(context.Context, domain.NewSupplierQuoteParams) (domain.SupplierQuote, error) {
	return domain.SupplierQuote{}, nil
}
func (s *testStorage) GetSupplierQuoteByID(context.Context, string) (*domain.SupplierQuote, error) {
	return nil, errors.New("unused")
}
func (s *testStorage) GetSupplierQuotes(context.Context, string) ([]domain.SupplierQuote, error) {
	return nil, nil
}
func (s *testStorage) UpdateSupplierQuoteStatus(context.Context, string, domain.SupplierQuoteStatus) (*domain.SupplierQuote, error) {
	return nil, nil
}
func (s *testStorage) CreateSupplierInvoice(context.Context, domain.NewSupplierInvoiceParams) (domain.SupplierInvoice, error) {
	return domain.SupplierInvoice{}, nil
}
func (s *testStorage) GetSupplierInvoiceByID(context.Context, string) (*domain.SupplierInvoice, error) {
	v := s.invoice
	return &v, nil
}
func (s *testStorage) GetSupplierInvoices(context.Context, string) ([]domain.SupplierInvoice, error) {
	return nil, nil
}
func (s *testStorage) UpdateSupplierInvoiceStatus(_ context.Context, _ string, status domain.SupplierInvoiceStatus) (*domain.SupplierInvoice, error) {
	s.updatedInvoiceStatus = status
	v := s.invoice
	v.Status = status
	return &v, nil
}
func (s *testStorage) CreateSupplierPayment(context.Context, domain.NewSupplierPaymentParams) (domain.SupplierPayment, error) {
	return domain.SupplierPayment{}, nil
}
func (s *testStorage) GetSupplierPaymentByID(context.Context, string) (*domain.SupplierPayment, error) {
	v := s.payment
	return &v, nil
}
func (s *testStorage) GetSupplierPayments(context.Context, string) ([]domain.SupplierPayment, error) {
	return nil, nil
}
func (s *testStorage) UpdateSupplierPaymentStatus(_ context.Context, _ string, status domain.SupplierPaymentStatus) (*domain.SupplierPayment, error) {
	s.updatedPaymentStatus = status
	v := s.payment
	v.Status = status
	return &v, nil
}
func (s *testStorage) GetSupplierOutstandingBalance(context.Context, string, domain.Currency) (*domain.SupplierOutstandingBalance, error) {
	return &domain.SupplierOutstandingBalance{TotalOutstanding: "12.34"}, nil
}
func (s *testStorage) GetMonthlyVATPosition(_ context.Context, month string) (*domain.MonthlyVATPosition, error) {
	s.vatCalls++
	v, err := domain.NewMonthlyVATPosition(month, []domain.VATPositionLine{{TaxRate: "21", SalesDebit: "21", PurchaseCredit: "10"}})
	return &v, err
}

func TestIssueSupplierInvoiceEnforcesLifecycleBeforePersisting(t *testing.T) {
	storage := &testStorage{invoice: domain.SupplierInvoice{Status: domain.SupplierInvoiceDraft}}
	result, err := NewService(storage).IssueSupplierInvoice(context.Background(), "invoice")
	require.NoError(t, err)
	require.Equal(t, domain.SupplierInvoiceIssued, storage.updatedInvoiceStatus)
	require.Equal(t, domain.SupplierInvoiceIssued, result.Status)
}

func TestPostSupplierPaymentEnforcesLifecycleBeforePersisting(t *testing.T) {
	storage := &testStorage{payment: domain.SupplierPayment{Status: domain.SupplierPaymentDraft}}
	_, err := NewService(storage).PostSupplierPayment(context.Background(), "payment")
	require.NoError(t, err)
	require.Equal(t, domain.SupplierPaymentPosted, storage.updatedPaymentStatus)
}

func TestGetMonthlyVATPositionRejectsInvalidMonthWithoutStorageCall(t *testing.T) {
	storage := &testStorage{}
	_, err := NewService(storage).GetMonthlyVATPosition(context.Background(), "2026-13")
	require.Error(t, err)
	require.Zero(t, storage.vatCalls)
}
