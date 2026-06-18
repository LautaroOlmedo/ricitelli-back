package server

import (
	"context"

	salesadministrationpb "ricitelli-back/cmd/http/gen/sales_administration"
	customer_receipt "ricitelli-back/internal/domain/customer-receipt"
	"ricitelli-back/internal/domain/remittance"
	sales_invoice "ricitelli-back/internal/domain/sales-invoice"
	sales_administration "ricitelli-back/internal/service/sales-administration"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SalesAdministrationServer is deliberately separate from the legacy combined
// server so integration ownership can register it from cmd/main.go later.
type SalesAdministrationServer struct {
	salesadministrationpb.UnimplementedSalesAdministrationServiceServer
	Service *sales_administration.Service
}

func NewSalesAdministrationServer(service *sales_administration.Service) *SalesAdministrationServer {
	return &SalesAdministrationServer{Service: service}
}

func (s *SalesAdministrationServer) CreateSalesInvoice(ctx context.Context, req *salesadministrationpb.CreateSalesInvoiceRequest) (*salesadministrationpb.SalesInvoice, error) {
	items := make([]sales_invoice.Item, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, sales_invoice.Item{
			LineNumber: item.LineNumber, ProductID: item.ProductId, Description: item.Description, Quantity: item.Quantity,
			UnitPrice: item.UnitPrice, TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount,
			TotalAmount: item.TotalAmount,
		})
	}
	invoice, err := s.Service.CreateSalesInvoice(ctx, sales_invoice.NewSalesInvoiceParams{
		CustomerID: req.CustomerId, SaleOrderID: req.SaleOrderId, DocumentType: sales_invoice.DocumentType(req.DocumentType),
		PointOfSale: req.PointOfSale, DocumentNumber: req.DocumentNumber, IssueDate: req.IssueDate, DueDate: req.DueDate,
		Currency: req.Currency, ExchangeRate: req.ExchangeRate, Subtotal: req.Subtotal, TaxTotal: req.TaxTotal,
		TotalAmount: req.TotalAmount, IdempotencyKey: req.IdempotencyKey, Items: items,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoSalesInvoice(invoice), nil
}

func (s *SalesAdministrationServer) GetSalesInvoiceByID(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.SalesInvoice, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	invoice, err := s.Service.GetSalesInvoiceByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoSalesInvoice(invoice), nil
}

func (s *SalesAdministrationServer) ListSalesInvoices(ctx context.Context, req *salesadministrationpb.ListSalesAdministrationRequest) (*salesadministrationpb.SalesInvoiceList, error) {
	invoices, err := s.Service.ListSalesInvoices(ctx, sales_administration.SalesInvoiceFilter{
		CustomerID: req.CustomerId, SaleOrderID: req.SaleOrderId, Status: req.Status,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := &salesadministrationpb.SalesInvoiceList{Invoices: make([]*salesadministrationpb.SalesInvoice, 0, len(invoices))}
	for i := range invoices {
		result.Invoices = append(result.Invoices, toProtoSalesInvoice(&invoices[i]))
	}
	return result, nil
}

func (s *SalesAdministrationServer) IssueSalesInvoice(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.SalesInvoice, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	invoice, err := s.Service.IssueSalesInvoice(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoSalesInvoice(invoice), nil
}

func (s *SalesAdministrationServer) GetSalesInvoiceOutstandingBalance(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.InvoiceOutstandingBalance, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	balance, err := s.Service.GetSalesInvoiceOutstandingBalance(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return &salesadministrationpb.InvoiceOutstandingBalance{SalesInvoiceId: req.Id, OutstandingBalance: balance}, nil
}

func (s *SalesAdministrationServer) CreateRemittance(ctx context.Context, req *salesadministrationpb.CreateRemittanceRequest) (*salesadministrationpb.Remittance, error) {
	items := make([]remittance.Item, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, remittance.Item{
			LineNumber: item.LineNumber, ProductID: item.ProductId, Description: item.Description,
			Quantity: item.Quantity, LotNumber: item.LotNumber,
		})
	}
	document, err := s.Service.CreateRemittance(ctx, remittance.NewRemittanceParams{
		CustomerID: req.CustomerId, SaleOrderID: req.SaleOrderId, SalesInvoiceID: req.SalesInvoiceId,
		PointOfSale: req.PointOfSale, DocumentNumber: req.DocumentNumber, IssueDate: req.IssueDate,
		DeliveryDate: req.DeliveryDate, IdempotencyKey: req.IdempotencyKey, Items: items,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoRemittance(document), nil
}

func (s *SalesAdministrationServer) GetRemittanceByID(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.Remittance, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	document, err := s.Service.GetRemittanceByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoRemittance(document), nil
}

func (s *SalesAdministrationServer) ListRemittances(ctx context.Context, req *salesadministrationpb.ListSalesAdministrationRequest) (*salesadministrationpb.RemittanceList, error) {
	documents, err := s.Service.ListRemittances(ctx, sales_administration.RemittanceFilter{
		CustomerID: req.CustomerId, SaleOrderID: req.SaleOrderId, Status: req.Status,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := &salesadministrationpb.RemittanceList{Remittances: make([]*salesadministrationpb.Remittance, 0, len(documents))}
	for i := range documents {
		result.Remittances = append(result.Remittances, toProtoRemittance(&documents[i]))
	}
	return result, nil
}

func (s *SalesAdministrationServer) ConfirmRemittance(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.Remittance, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	document, err := s.Service.ConfirmRemittance(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoRemittance(document), nil
}

func (s *SalesAdministrationServer) CreateCustomerReceipt(ctx context.Context, req *salesadministrationpb.CreateCustomerReceiptRequest) (*salesadministrationpb.CustomerReceipt, error) {
	allocations := make([]customer_receipt.Allocation, 0, len(req.Allocations))
	for _, allocation := range req.Allocations {
		allocations = append(allocations, customer_receipt.Allocation{
			SalesInvoiceID: allocation.SalesInvoiceId, AllocatedAmount: allocation.AllocatedAmount,
		})
	}
	receipt, err := s.Service.CreateCustomerReceipt(ctx, customer_receipt.NewCustomerReceiptParams{
		CustomerID: req.CustomerId, ReceiptNumber: req.ReceiptNumber, ReceiptDate: req.ReceiptDate,
		Currency: req.Currency, ExchangeRate: req.ExchangeRate, Amount: req.Amount, PaymentMethod: req.PaymentMethod,
		PaymentReference: req.PaymentReference, IdempotencyKey: req.IdempotencyKey, Allocations: allocations,
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoCustomerReceipt(receipt), nil
}

func (s *SalesAdministrationServer) GetCustomerReceiptByID(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.CustomerReceipt, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	receipt, err := s.Service.GetCustomerReceiptByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	return toProtoCustomerReceipt(receipt), nil
}

func (s *SalesAdministrationServer) ListCustomerReceipts(ctx context.Context, req *salesadministrationpb.ListSalesAdministrationRequest) (*salesadministrationpb.CustomerReceiptList, error) {
	receipts, err := s.Service.ListCustomerReceipts(ctx, sales_administration.CustomerReceiptFilter{
		CustomerID: req.CustomerId, SaleOrderID: req.SaleOrderId, Status: req.Status,
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := &salesadministrationpb.CustomerReceiptList{Receipts: make([]*salesadministrationpb.CustomerReceipt, 0, len(receipts))}
	for i := range receipts {
		result.Receipts = append(result.Receipts, toProtoCustomerReceipt(&receipts[i]))
	}
	return result, nil
}

func (s *SalesAdministrationServer) PostCustomerReceipt(ctx context.Context, req *salesadministrationpb.IDRequest) (*salesadministrationpb.CustomerReceipt, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	receipt, err := s.Service.PostCustomerReceipt(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoCustomerReceipt(receipt), nil
}

func toProtoSalesInvoice(invoice *sales_invoice.SalesInvoice) *salesadministrationpb.SalesInvoice {
	items := make([]*salesadministrationpb.SalesInvoiceItem, 0, len(invoice.GetItems()))
	for _, item := range invoice.GetItems() {
		items = append(items, &salesadministrationpb.SalesInvoiceItem{
			LineNumber: item.LineNumber, ProductId: item.ProductID, Description: item.Description, Quantity: item.Quantity,
			UnitPrice: item.UnitPrice, TaxRate: item.TaxRate, NetAmount: item.NetAmount, TaxAmount: item.TaxAmount,
			TotalAmount: item.TotalAmount,
		})
	}
	return &salesadministrationpb.SalesInvoice{
		Id: invoice.GetID(), CustomerId: invoice.GetCustomerID(), SaleOrderId: invoice.GetSaleOrderID(),
		DocumentType: string(invoice.GetDocumentType()), PointOfSale: invoice.GetPointOfSale(),
		DocumentNumber: invoice.GetDocumentNumber(), IssueDate: invoice.GetIssueDate(), DueDate: invoice.GetDueDate(),
		Currency: invoice.GetCurrency(), ExchangeRate: invoice.GetExchangeRate(), Subtotal: invoice.GetSubtotal(),
		TaxTotal: invoice.GetTaxTotal(), TotalAmount: invoice.GetTotalAmount(), Status: string(invoice.GetStatus()),
		IdempotencyKey: invoice.GetIdempotencyKey(), Items: items, CreatedAt: invoice.GetCreatedAt(), UpdatedAt: invoice.GetUpdatedAt(),
	}
}

func toProtoRemittance(document *remittance.Remittance) *salesadministrationpb.Remittance {
	items := make([]*salesadministrationpb.RemittanceItem, 0, len(document.GetItems()))
	for _, item := range document.GetItems() {
		items = append(items, &salesadministrationpb.RemittanceItem{
			LineNumber: item.LineNumber, ProductId: item.ProductID, Description: item.Description,
			Quantity: item.Quantity, LotNumber: item.LotNumber,
		})
	}
	return &salesadministrationpb.Remittance{
		Id: document.GetID(), CustomerId: document.GetCustomerID(), SaleOrderId: document.GetSaleOrderID(),
		SalesInvoiceId: document.GetSalesInvoiceID(), PointOfSale: document.GetPointOfSale(),
		DocumentNumber: document.GetDocumentNumber(), IssueDate: document.GetIssueDate(), DeliveryDate: document.GetDeliveryDate(),
		Status: string(document.GetStatus()), IdempotencyKey: document.GetIdempotencyKey(), Items: items,
		CreatedAt: document.GetCreatedAt(), UpdatedAt: document.GetUpdatedAt(),
	}
}

func toProtoCustomerReceipt(receipt *customer_receipt.CustomerReceipt) *salesadministrationpb.CustomerReceipt {
	allocations := make([]*salesadministrationpb.CustomerReceiptAllocation, 0, len(receipt.GetAllocations()))
	for _, allocation := range receipt.GetAllocations() {
		allocations = append(allocations, &salesadministrationpb.CustomerReceiptAllocation{
			SalesInvoiceId: allocation.SalesInvoiceID, AllocatedAmount: allocation.AllocatedAmount,
		})
	}
	return &salesadministrationpb.CustomerReceipt{
		Id: receipt.GetID(), CustomerId: receipt.GetCustomerID(), ReceiptNumber: receipt.GetReceiptNumber(),
		ReceiptDate: receipt.GetReceiptDate(), Currency: receipt.GetCurrency(), ExchangeRate: receipt.GetExchangeRate(),
		Amount: receipt.GetAmount(), PaymentMethod: receipt.GetPaymentMethod(), PaymentReference: receipt.GetPaymentReference(),
		Status: string(receipt.GetStatus()), IdempotencyKey: receipt.GetIdempotencyKey(), Allocations: allocations,
		CreatedAt: receipt.GetCreatedAt(), UpdatedAt: receipt.GetUpdatedAt(),
	}
}
