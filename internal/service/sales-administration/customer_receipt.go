package sales_administration

import (
	"context"

	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
)

func (s *Service) CreateCustomerReceipt(ctx context.Context, params customer_receipt.NewCustomerReceiptParams) (*customer_receipt.CustomerReceipt, error) {
	receipt, err := customer_receipt.NewCustomerReceipt(params)
	if err != nil {
		return nil, err
	}
	return s.Storage.SaveCustomerReceipt(ctx, receipt)
}

func (s *Service) GetCustomerReceiptByID(ctx context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	return s.Storage.GetCustomerReceiptByID(ctx, id)
}

func (s *Service) ListCustomerReceipts(ctx context.Context, filter CustomerReceiptFilter) ([]customer_receipt.CustomerReceipt, error) {
	return s.Storage.ListCustomerReceipts(ctx, filter)
}

func (s *Service) PostCustomerReceipt(ctx context.Context, id string) (*customer_receipt.CustomerReceipt, error) {
	return s.Storage.PostCustomerReceipt(ctx, id)
}
