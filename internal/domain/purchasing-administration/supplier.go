package purchasing_administration

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Supplier struct {
	ID             string
	TaxID          string
	SocialReason   string
	TradeName      string
	Email          string
	Phone          string
	Address        string
	Active         bool
	IdempotencyKey string
	CreatedAt      string
	UpdatedAt      string
}

type NewSupplierParams struct {
	TaxID, SocialReason, TradeName, Email, Phone, Address, IdempotencyKey string
}

type UpdateSupplierParams struct {
	TaxID, SocialReason, TradeName, Email, Phone, Address string
}

func NewSupplier(params NewSupplierParams) (Supplier, error) {
	if strings.TrimSpace(params.TaxID) == "" {
		return Supplier{}, errors.New("tax_id is required")
	}
	if strings.TrimSpace(params.SocialReason) == "" {
		return Supplier{}, errors.New("social_reason is required")
	}
	if strings.TrimSpace(params.IdempotencyKey) == "" {
		return Supplier{}, errors.New("idempotency_key is required")
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	return Supplier{
		ID: uuid.NewString(), TaxID: params.TaxID, SocialReason: params.SocialReason,
		TradeName: params.TradeName, Email: params.Email, Phone: params.Phone, Address: params.Address,
		Active: true, IdempotencyKey: params.IdempotencyKey, CreatedAt: now, UpdatedAt: now,
	}, nil
}

func (s *Supplier) Deactivate() error {
	if !s.Active {
		return errors.New("supplier is already inactive")
	}
	s.Active = false
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
	return nil
}

func (s *Supplier) Update(params UpdateSupplierParams) {
	if params.TaxID != "" {
		s.TaxID = params.TaxID
	}
	if params.SocialReason != "" {
		s.SocialReason = params.SocialReason
	}
	if params.TradeName != "" {
		s.TradeName = params.TradeName
	}
	if params.Email != "" {
		s.Email = params.Email
	}
	if params.Phone != "" {
		s.Phone = params.Phone
	}
	if params.Address != "" {
		s.Address = params.Address
	}
	s.UpdatedAt = time.Now().UTC().Format(time.RFC3339Nano)
}
