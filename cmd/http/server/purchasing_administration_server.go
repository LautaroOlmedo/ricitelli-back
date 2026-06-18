package server

import (
	"context"

	purchasingpb "ricitelli-back/cmd/http/gen/purchasing_administration"
	domain "ricitelli-back/internal/domain/purchasing-administration"
	purchasing_svc "ricitelli-back/internal/service/purchasing-administration"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// PurchasingAdministrationServer is intentionally separate from Server so integration can be owned independently.
type PurchasingAdministrationServer struct {
	purchasingpb.UnimplementedPurchasingAdministrationServiceServer
	Svc *purchasing_svc.Service
}

func NewPurchasingAdministrationServer(svc *purchasing_svc.Service) *PurchasingAdministrationServer {
	return &PurchasingAdministrationServer{Svc: svc}
}

func requireID(id string) error {
	if id == "" {
		return status.Error(codes.InvalidArgument, "id is required")
	}
	return nil
}

func invalidArgument(err error) error { return status.Error(codes.InvalidArgument, err.Error()) }
func notFound(err error) error        { return status.Error(codes.NotFound, err.Error()) }
func internal(err error) error        { return status.Error(codes.Internal, err.Error()) }

func (s *PurchasingAdministrationServer) CreateSupplier(ctx context.Context, req *purchasingpb.CreateSupplierRequest) (*purchasingpb.Supplier, error) {
	value, err := s.Svc.CreateSupplier(ctx, domain.NewSupplierParams{TaxID: req.TaxId, SocialReason: req.SocialReason,
		TradeName: req.TradeName, Email: req.Email, Phone: req.Phone, Address: req.Address, IdempotencyKey: req.IdempotencyKey})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplier(&value), nil
}

func (s *PurchasingAdministrationServer) GetSupplier(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.Supplier, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.GetSupplierByID(ctx, req.Id)
	if err != nil {
		return nil, notFound(err)
	}
	return toProtoSupplier(value), nil
}

func (s *PurchasingAdministrationServer) ListSuppliers(ctx context.Context, req *purchasingpb.SupplierListRequest) (*purchasingpb.SupplierListResponse, error) {
	values, err := s.Svc.GetSuppliers(ctx, req.IncludeInactive)
	if err != nil {
		return nil, internal(err)
	}
	result := make([]*purchasingpb.Supplier, len(values))
	for i := range values {
		result[i] = toProtoSupplier(&values[i])
	}
	return &purchasingpb.SupplierListResponse{Suppliers: result}, nil
}

func (s *PurchasingAdministrationServer) UpdateSupplier(ctx context.Context, req *purchasingpb.UpdateSupplierRequest) (*purchasingpb.Supplier, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.UpdateSupplier(ctx, req.Id, domain.UpdateSupplierParams{
		TaxID: req.TaxId, SocialReason: req.SocialReason, TradeName: req.TradeName,
		Email: req.Email, Phone: req.Phone, Address: req.Address,
	})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplier(value), nil
}

func (s *PurchasingAdministrationServer) DeactivateSupplier(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.Supplier, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.DeactivateSupplier(ctx, req.Id)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplier(value), nil
}

func (s *PurchasingAdministrationServer) CreatePurchaseNeed(ctx context.Context, req *purchasingpb.CreatePurchaseNeedRequest) (*purchasingpb.PurchaseNeed, error) {
	value, err := s.Svc.CreatePurchaseNeed(ctx, domain.NewPurchaseNeedParams{NeedNumber: req.NeedNumber,
		RequestedDate: req.RequestedDate, RequiredByDate: req.RequiredByDate, Notes: req.Notes,
		IdempotencyKey: req.IdempotencyKey, Items: purchaseNeedItemsFromProto(req.Items)})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoPurchaseNeed(&value), nil
}

func (s *PurchasingAdministrationServer) GetPurchaseNeed(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.PurchaseNeed, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.GetPurchaseNeedByID(ctx, req.Id)
	if err != nil {
		return nil, notFound(err)
	}
	return toProtoPurchaseNeed(value), nil
}

func (s *PurchasingAdministrationServer) ListPurchaseNeeds(ctx context.Context, _ *purchasingpb.PurchaseNeedListRequest) (*purchasingpb.PurchaseNeedListResponse, error) {
	values, err := s.Svc.GetPurchaseNeeds(ctx)
	if err != nil {
		return nil, internal(err)
	}
	result := make([]*purchasingpb.PurchaseNeed, len(values))
	for i := range values {
		result[i] = toProtoPurchaseNeed(&values[i])
	}
	return &purchasingpb.PurchaseNeedListResponse{PurchaseNeeds: result}, nil
}

func (s *PurchasingAdministrationServer) UpdatePurchaseNeedStatus(ctx context.Context, req *purchasingpb.StatusRequest) (*purchasingpb.PurchaseNeed, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.UpdatePurchaseNeedStatus(ctx, req.Id, domain.PurchaseNeedStatus(req.Status))
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoPurchaseNeed(value), nil
}

func (s *PurchasingAdministrationServer) CreateSupplierQuote(ctx context.Context, req *purchasingpb.CreateSupplierQuoteRequest) (*purchasingpb.SupplierQuote, error) {
	value, err := s.Svc.CreateSupplierQuote(ctx, domain.NewSupplierQuoteParams{SupplierID: req.SupplierId,
		PurchaseNeedID: req.PurchaseNeedId, QuoteNumber: req.QuoteNumber, QuoteDate: req.QuoteDate, ValidUntil: req.ValidUntil,
		Currency: domain.Currency(req.Currency), ExchangeRate: req.ExchangeRate, Subtotal: req.Subtotal, TaxTotal: req.TaxTotal,
		TotalAmount: req.TotalAmount, IdempotencyKey: req.IdempotencyKey, Items: quoteItemsFromProto(req.Items)})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierQuote(&value), nil
}

func (s *PurchasingAdministrationServer) GetSupplierQuote(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierQuote, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.GetSupplierQuoteByID(ctx, req.Id)
	if err != nil {
		return nil, notFound(err)
	}
	return toProtoSupplierQuote(value), nil
}

func (s *PurchasingAdministrationServer) ListSupplierQuotes(ctx context.Context, req *purchasingpb.SupplierScopedRequest) (*purchasingpb.SupplierQuoteListResponse, error) {
	values, err := s.Svc.GetSupplierQuotes(ctx, req.SupplierId)
	if err != nil {
		return nil, internal(err)
	}
	result := make([]*purchasingpb.SupplierQuote, len(values))
	for i := range values {
		result[i] = toProtoSupplierQuote(&values[i])
	}
	return &purchasingpb.SupplierQuoteListResponse{SupplierQuotes: result}, nil
}

func (s *PurchasingAdministrationServer) UpdateSupplierQuoteStatus(ctx context.Context, req *purchasingpb.StatusRequest) (*purchasingpb.SupplierQuote, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.UpdateSupplierQuoteStatus(ctx, req.Id, domain.SupplierQuoteStatus(req.Status))
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierQuote(value), nil
}

func (s *PurchasingAdministrationServer) CreateSupplierInvoice(ctx context.Context, req *purchasingpb.CreateSupplierInvoiceRequest) (*purchasingpb.SupplierInvoice, error) {
	value, err := s.Svc.CreateSupplierInvoice(ctx, domain.NewSupplierInvoiceParams{SupplierID: req.SupplierId,
		SupplierQuoteID: req.SupplierQuoteId, DocumentType: domain.SupplierInvoiceDocumentType(req.DocumentType),
		PointOfSale: req.PointOfSale, DocumentNumber: req.DocumentNumber, IssueDate: req.IssueDate, DueDate: req.DueDate,
		Currency: domain.Currency(req.Currency), ExchangeRate: req.ExchangeRate, Subtotal: req.Subtotal, TaxTotal: req.TaxTotal,
		TotalAmount: req.TotalAmount, IdempotencyKey: req.IdempotencyKey, Items: invoiceItemsFromProto(req.Items)})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierInvoice(&value), nil
}

func (s *PurchasingAdministrationServer) GetSupplierInvoice(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierInvoice, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.GetSupplierInvoiceByID(ctx, req.Id)
	if err != nil {
		return nil, notFound(err)
	}
	return toProtoSupplierInvoice(value), nil
}

func (s *PurchasingAdministrationServer) ListSupplierInvoices(ctx context.Context, req *purchasingpb.SupplierScopedRequest) (*purchasingpb.SupplierInvoiceListResponse, error) {
	values, err := s.Svc.GetSupplierInvoices(ctx, req.SupplierId)
	if err != nil {
		return nil, internal(err)
	}
	result := make([]*purchasingpb.SupplierInvoice, len(values))
	for i := range values {
		result[i] = toProtoSupplierInvoice(&values[i])
	}
	return &purchasingpb.SupplierInvoiceListResponse{SupplierInvoices: result}, nil
}

func (s *PurchasingAdministrationServer) IssueSupplierInvoice(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierInvoice, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.IssueSupplierInvoice(ctx, req.Id)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierInvoice(value), nil
}

func (s *PurchasingAdministrationServer) VoidSupplierInvoice(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierInvoice, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.VoidSupplierInvoice(ctx, req.Id)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierInvoice(value), nil
}

func (s *PurchasingAdministrationServer) CreateSupplierPayment(ctx context.Context, req *purchasingpb.CreateSupplierPaymentRequest) (*purchasingpb.SupplierPayment, error) {
	value, err := s.Svc.CreateSupplierPayment(ctx, domain.NewSupplierPaymentParams{SupplierID: req.SupplierId,
		PaymentNumber: req.PaymentNumber, PaymentDate: req.PaymentDate, Currency: domain.Currency(req.Currency),
		ExchangeRate: req.ExchangeRate, Amount: req.Amount, PaymentMethod: req.PaymentMethod,
		PaymentReference: req.PaymentReference, IdempotencyKey: req.IdempotencyKey,
		Allocations: paymentAllocationsFromProto(req.Allocations)})
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierPayment(&value), nil
}

func (s *PurchasingAdministrationServer) GetSupplierPayment(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierPayment, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.GetSupplierPaymentByID(ctx, req.Id)
	if err != nil {
		return nil, notFound(err)
	}
	return toProtoSupplierPayment(value), nil
}

func (s *PurchasingAdministrationServer) ListSupplierPayments(ctx context.Context, req *purchasingpb.SupplierScopedRequest) (*purchasingpb.SupplierPaymentListResponse, error) {
	values, err := s.Svc.GetSupplierPayments(ctx, req.SupplierId)
	if err != nil {
		return nil, internal(err)
	}
	result := make([]*purchasingpb.SupplierPayment, len(values))
	for i := range values {
		result[i] = toProtoSupplierPayment(&values[i])
	}
	return &purchasingpb.SupplierPaymentListResponse{SupplierPayments: result}, nil
}

func (s *PurchasingAdministrationServer) PostSupplierPayment(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierPayment, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.PostSupplierPayment(ctx, req.Id)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierPayment(value), nil
}

func (s *PurchasingAdministrationServer) VoidSupplierPayment(ctx context.Context, req *purchasingpb.IDRequest) (*purchasingpb.SupplierPayment, error) {
	if err := requireID(req.Id); err != nil {
		return nil, err
	}
	value, err := s.Svc.VoidSupplierPayment(ctx, req.Id)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoSupplierPayment(value), nil
}

func (s *PurchasingAdministrationServer) GetSupplierOutstandingBalance(ctx context.Context, req *purchasingpb.OutstandingBalanceRequest) (*purchasingpb.SupplierOutstandingBalance, error) {
	if req.SupplierId == "" {
		return nil, status.Error(codes.InvalidArgument, "supplier_id is required")
	}
	value, err := s.Svc.GetSupplierOutstandingBalance(ctx, req.SupplierId, domain.Currency(req.Currency))
	if err != nil {
		return nil, internal(err)
	}
	return toProtoOutstanding(value), nil
}

func (s *PurchasingAdministrationServer) GetMonthlyVATPosition(ctx context.Context, req *purchasingpb.MonthlyVATPositionRequest) (*purchasingpb.MonthlyVATPosition, error) {
	value, err := s.Svc.GetMonthlyVATPosition(ctx, req.Month)
	if err != nil {
		return nil, invalidArgument(err)
	}
	return toProtoVATPosition(value), nil
}
