package sales_administration

import (
	"context"

	"ricitelli-back/internal/domain/remittance"
)

func (s *Service) CreateRemittance(ctx context.Context, params remittance.NewRemittanceParams) (*remittance.Remittance, error) {
	document, err := remittance.NewRemittance(params)
	if err != nil {
		return nil, err
	}
	return s.Storage.SaveRemittance(ctx, document)
}

func (s *Service) GetRemittanceByID(ctx context.Context, id string) (*remittance.Remittance, error) {
	return s.Storage.GetRemittanceByID(ctx, id)
}

func (s *Service) ListRemittances(ctx context.Context, filter RemittanceFilter) ([]remittance.Remittance, error) {
	return s.Storage.ListRemittances(ctx, filter)
}

// ConfirmRemittance delegates the atomic idempotent confirmation and physical
// dispatch to storage. The dispatch reference is always the remittance ID.
func (s *Service) ConfirmRemittance(ctx context.Context, id string) (*remittance.Remittance, error) {
	return s.Storage.ConfirmRemittance(ctx, id, userIDFromContext(ctx))
}
