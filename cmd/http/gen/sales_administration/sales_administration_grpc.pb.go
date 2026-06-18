// Code generated from sales_administration.proto. DO NOT EDIT.

package salesadministrationpb

import (
	context "context"

	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const _ = grpc.SupportPackageIsVersion9

const (
	SalesAdministrationService_CreateSalesInvoice_FullMethodName                = "/sales_administration.SalesAdministrationService/CreateSalesInvoice"
	SalesAdministrationService_GetSalesInvoiceByID_FullMethodName               = "/sales_administration.SalesAdministrationService/GetSalesInvoiceByID"
	SalesAdministrationService_ListSalesInvoices_FullMethodName                 = "/sales_administration.SalesAdministrationService/ListSalesInvoices"
	SalesAdministrationService_IssueSalesInvoice_FullMethodName                 = "/sales_administration.SalesAdministrationService/IssueSalesInvoice"
	SalesAdministrationService_GetSalesInvoiceOutstandingBalance_FullMethodName = "/sales_administration.SalesAdministrationService/GetSalesInvoiceOutstandingBalance"
	SalesAdministrationService_CreateRemittance_FullMethodName                  = "/sales_administration.SalesAdministrationService/CreateRemittance"
	SalesAdministrationService_GetRemittanceByID_FullMethodName                 = "/sales_administration.SalesAdministrationService/GetRemittanceByID"
	SalesAdministrationService_ListRemittances_FullMethodName                   = "/sales_administration.SalesAdministrationService/ListRemittances"
	SalesAdministrationService_ConfirmRemittance_FullMethodName                 = "/sales_administration.SalesAdministrationService/ConfirmRemittance"
	SalesAdministrationService_CreateCustomerReceipt_FullMethodName             = "/sales_administration.SalesAdministrationService/CreateCustomerReceipt"
	SalesAdministrationService_GetCustomerReceiptByID_FullMethodName            = "/sales_administration.SalesAdministrationService/GetCustomerReceiptByID"
	SalesAdministrationService_ListCustomerReceipts_FullMethodName              = "/sales_administration.SalesAdministrationService/ListCustomerReceipts"
	SalesAdministrationService_PostCustomerReceipt_FullMethodName               = "/sales_administration.SalesAdministrationService/PostCustomerReceipt"
)

type SalesAdministrationServiceClient interface {
	CreateSalesInvoice(context.Context, *CreateSalesInvoiceRequest, ...grpc.CallOption) (*SalesInvoice, error)
	GetSalesInvoiceByID(context.Context, *IDRequest, ...grpc.CallOption) (*SalesInvoice, error)
	ListSalesInvoices(context.Context, *ListSalesAdministrationRequest, ...grpc.CallOption) (*SalesInvoiceList, error)
	IssueSalesInvoice(context.Context, *IDRequest, ...grpc.CallOption) (*SalesInvoice, error)
	GetSalesInvoiceOutstandingBalance(context.Context, *IDRequest, ...grpc.CallOption) (*InvoiceOutstandingBalance, error)
	CreateRemittance(context.Context, *CreateRemittanceRequest, ...grpc.CallOption) (*Remittance, error)
	GetRemittanceByID(context.Context, *IDRequest, ...grpc.CallOption) (*Remittance, error)
	ListRemittances(context.Context, *ListSalesAdministrationRequest, ...grpc.CallOption) (*RemittanceList, error)
	ConfirmRemittance(context.Context, *IDRequest, ...grpc.CallOption) (*Remittance, error)
	CreateCustomerReceipt(context.Context, *CreateCustomerReceiptRequest, ...grpc.CallOption) (*CustomerReceipt, error)
	GetCustomerReceiptByID(context.Context, *IDRequest, ...grpc.CallOption) (*CustomerReceipt, error)
	ListCustomerReceipts(context.Context, *ListSalesAdministrationRequest, ...grpc.CallOption) (*CustomerReceiptList, error)
	PostCustomerReceipt(context.Context, *IDRequest, ...grpc.CallOption) (*CustomerReceipt, error)
}

type salesAdministrationServiceClient struct{ cc grpc.ClientConnInterface }

func NewSalesAdministrationServiceClient(cc grpc.ClientConnInterface) SalesAdministrationServiceClient {
	return &salesAdministrationServiceClient{cc}
}

func invoke[Req any, Resp any](ctx context.Context, cc grpc.ClientConnInterface, method string, in *Req, opts ...grpc.CallOption) (*Resp, error) {
	out := new(Resp)
	err := cc.Invoke(ctx, method, in, out, append([]grpc.CallOption{grpc.StaticMethod()}, opts...)...)
	return out, err
}

func (c *salesAdministrationServiceClient) CreateSalesInvoice(ctx context.Context, in *CreateSalesInvoiceRequest, opts ...grpc.CallOption) (*SalesInvoice, error) {
	return invoke[CreateSalesInvoiceRequest, SalesInvoice](ctx, c.cc, SalesAdministrationService_CreateSalesInvoice_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) GetSalesInvoiceByID(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*SalesInvoice, error) {
	return invoke[IDRequest, SalesInvoice](ctx, c.cc, SalesAdministrationService_GetSalesInvoiceByID_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) ListSalesInvoices(ctx context.Context, in *ListSalesAdministrationRequest, opts ...grpc.CallOption) (*SalesInvoiceList, error) {
	return invoke[ListSalesAdministrationRequest, SalesInvoiceList](ctx, c.cc, SalesAdministrationService_ListSalesInvoices_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) IssueSalesInvoice(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*SalesInvoice, error) {
	return invoke[IDRequest, SalesInvoice](ctx, c.cc, SalesAdministrationService_IssueSalesInvoice_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) GetSalesInvoiceOutstandingBalance(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*InvoiceOutstandingBalance, error) {
	return invoke[IDRequest, InvoiceOutstandingBalance](ctx, c.cc, SalesAdministrationService_GetSalesInvoiceOutstandingBalance_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) CreateRemittance(ctx context.Context, in *CreateRemittanceRequest, opts ...grpc.CallOption) (*Remittance, error) {
	return invoke[CreateRemittanceRequest, Remittance](ctx, c.cc, SalesAdministrationService_CreateRemittance_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) GetRemittanceByID(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*Remittance, error) {
	return invoke[IDRequest, Remittance](ctx, c.cc, SalesAdministrationService_GetRemittanceByID_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) ListRemittances(ctx context.Context, in *ListSalesAdministrationRequest, opts ...grpc.CallOption) (*RemittanceList, error) {
	return invoke[ListSalesAdministrationRequest, RemittanceList](ctx, c.cc, SalesAdministrationService_ListRemittances_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) ConfirmRemittance(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*Remittance, error) {
	return invoke[IDRequest, Remittance](ctx, c.cc, SalesAdministrationService_ConfirmRemittance_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) CreateCustomerReceipt(ctx context.Context, in *CreateCustomerReceiptRequest, opts ...grpc.CallOption) (*CustomerReceipt, error) {
	return invoke[CreateCustomerReceiptRequest, CustomerReceipt](ctx, c.cc, SalesAdministrationService_CreateCustomerReceipt_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) GetCustomerReceiptByID(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*CustomerReceipt, error) {
	return invoke[IDRequest, CustomerReceipt](ctx, c.cc, SalesAdministrationService_GetCustomerReceiptByID_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) ListCustomerReceipts(ctx context.Context, in *ListSalesAdministrationRequest, opts ...grpc.CallOption) (*CustomerReceiptList, error) {
	return invoke[ListSalesAdministrationRequest, CustomerReceiptList](ctx, c.cc, SalesAdministrationService_ListCustomerReceipts_FullMethodName, in, opts...)
}
func (c *salesAdministrationServiceClient) PostCustomerReceipt(ctx context.Context, in *IDRequest, opts ...grpc.CallOption) (*CustomerReceipt, error) {
	return invoke[IDRequest, CustomerReceipt](ctx, c.cc, SalesAdministrationService_PostCustomerReceipt_FullMethodName, in, opts...)
}

type SalesAdministrationServiceServer interface {
	CreateSalesInvoice(context.Context, *CreateSalesInvoiceRequest) (*SalesInvoice, error)
	GetSalesInvoiceByID(context.Context, *IDRequest) (*SalesInvoice, error)
	ListSalesInvoices(context.Context, *ListSalesAdministrationRequest) (*SalesInvoiceList, error)
	IssueSalesInvoice(context.Context, *IDRequest) (*SalesInvoice, error)
	GetSalesInvoiceOutstandingBalance(context.Context, *IDRequest) (*InvoiceOutstandingBalance, error)
	CreateRemittance(context.Context, *CreateRemittanceRequest) (*Remittance, error)
	GetRemittanceByID(context.Context, *IDRequest) (*Remittance, error)
	ListRemittances(context.Context, *ListSalesAdministrationRequest) (*RemittanceList, error)
	ConfirmRemittance(context.Context, *IDRequest) (*Remittance, error)
	CreateCustomerReceipt(context.Context, *CreateCustomerReceiptRequest) (*CustomerReceipt, error)
	GetCustomerReceiptByID(context.Context, *IDRequest) (*CustomerReceipt, error)
	ListCustomerReceipts(context.Context, *ListSalesAdministrationRequest) (*CustomerReceiptList, error)
	PostCustomerReceipt(context.Context, *IDRequest) (*CustomerReceipt, error)
	mustEmbedUnimplementedSalesAdministrationServiceServer()
}

type UnimplementedSalesAdministrationServiceServer struct{}

func (UnimplementedSalesAdministrationServiceServer) CreateSalesInvoice(context.Context, *CreateSalesInvoiceRequest) (*SalesInvoice, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateSalesInvoice not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) GetSalesInvoiceByID(context.Context, *IDRequest) (*SalesInvoice, error) {
	return nil, status.Error(codes.Unimplemented, "method GetSalesInvoiceByID not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) ListSalesInvoices(context.Context, *ListSalesAdministrationRequest) (*SalesInvoiceList, error) {
	return nil, status.Error(codes.Unimplemented, "method ListSalesInvoices not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) IssueSalesInvoice(context.Context, *IDRequest) (*SalesInvoice, error) {
	return nil, status.Error(codes.Unimplemented, "method IssueSalesInvoice not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) GetSalesInvoiceOutstandingBalance(context.Context, *IDRequest) (*InvoiceOutstandingBalance, error) {
	return nil, status.Error(codes.Unimplemented, "method GetSalesInvoiceOutstandingBalance not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) CreateRemittance(context.Context, *CreateRemittanceRequest) (*Remittance, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateRemittance not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) GetRemittanceByID(context.Context, *IDRequest) (*Remittance, error) {
	return nil, status.Error(codes.Unimplemented, "method GetRemittanceByID not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) ListRemittances(context.Context, *ListSalesAdministrationRequest) (*RemittanceList, error) {
	return nil, status.Error(codes.Unimplemented, "method ListRemittances not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) ConfirmRemittance(context.Context, *IDRequest) (*Remittance, error) {
	return nil, status.Error(codes.Unimplemented, "method ConfirmRemittance not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) CreateCustomerReceipt(context.Context, *CreateCustomerReceiptRequest) (*CustomerReceipt, error) {
	return nil, status.Error(codes.Unimplemented, "method CreateCustomerReceipt not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) GetCustomerReceiptByID(context.Context, *IDRequest) (*CustomerReceipt, error) {
	return nil, status.Error(codes.Unimplemented, "method GetCustomerReceiptByID not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) ListCustomerReceipts(context.Context, *ListSalesAdministrationRequest) (*CustomerReceiptList, error) {
	return nil, status.Error(codes.Unimplemented, "method ListCustomerReceipts not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) PostCustomerReceipt(context.Context, *IDRequest) (*CustomerReceipt, error) {
	return nil, status.Error(codes.Unimplemented, "method PostCustomerReceipt not implemented")
}
func (UnimplementedSalesAdministrationServiceServer) mustEmbedUnimplementedSalesAdministrationServiceServer() {
}

func RegisterSalesAdministrationServiceServer(s grpc.ServiceRegistrar, srv SalesAdministrationServiceServer) {
	s.RegisterService(&SalesAdministrationService_ServiceDesc, srv)
}

func unaryHandler[Req any, Resp any](srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor, method string, call func(SalesAdministrationServiceServer, context.Context, *Req) (*Resp, error)) (interface{}, error) {
	in := new(Req)
	if err := dec(in); err != nil {
		return nil, err
	}
	server := srv.(SalesAdministrationServiceServer)
	if interceptor == nil {
		return call(server, ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: method}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) { return call(server, ctx, req.(*Req)) }
	return interceptor(ctx, in, info, handler)
}

func _SalesAdministrationService_CreateSalesInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_CreateSalesInvoice_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *CreateSalesInvoiceRequest) (*SalesInvoice, error) {
		return s.CreateSalesInvoice(c, r)
	})
}
func _SalesAdministrationService_GetSalesInvoiceByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_GetSalesInvoiceByID_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*SalesInvoice, error) {
		return s.GetSalesInvoiceByID(c, r)
	})
}
func _SalesAdministrationService_ListSalesInvoices_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_ListSalesInvoices_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *ListSalesAdministrationRequest) (*SalesInvoiceList, error) {
		return s.ListSalesInvoices(c, r)
	})
}
func _SalesAdministrationService_IssueSalesInvoice_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_IssueSalesInvoice_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*SalesInvoice, error) {
		return s.IssueSalesInvoice(c, r)
	})
}
func _SalesAdministrationService_GetSalesInvoiceOutstandingBalance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_GetSalesInvoiceOutstandingBalance_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*InvoiceOutstandingBalance, error) {
		return s.GetSalesInvoiceOutstandingBalance(c, r)
	})
}
func _SalesAdministrationService_CreateRemittance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_CreateRemittance_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *CreateRemittanceRequest) (*Remittance, error) {
		return s.CreateRemittance(c, r)
	})
}
func _SalesAdministrationService_GetRemittanceByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_GetRemittanceByID_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*Remittance, error) {
		return s.GetRemittanceByID(c, r)
	})
}
func _SalesAdministrationService_ListRemittances_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_ListRemittances_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *ListSalesAdministrationRequest) (*RemittanceList, error) {
		return s.ListRemittances(c, r)
	})
}
func _SalesAdministrationService_ConfirmRemittance_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_ConfirmRemittance_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*Remittance, error) {
		return s.ConfirmRemittance(c, r)
	})
}
func _SalesAdministrationService_CreateCustomerReceipt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_CreateCustomerReceipt_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *CreateCustomerReceiptRequest) (*CustomerReceipt, error) {
		return s.CreateCustomerReceipt(c, r)
	})
}
func _SalesAdministrationService_GetCustomerReceiptByID_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_GetCustomerReceiptByID_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*CustomerReceipt, error) {
		return s.GetCustomerReceiptByID(c, r)
	})
}
func _SalesAdministrationService_ListCustomerReceipts_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_ListCustomerReceipts_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *ListSalesAdministrationRequest) (*CustomerReceiptList, error) {
		return s.ListCustomerReceipts(c, r)
	})
}
func _SalesAdministrationService_PostCustomerReceipt_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	return unaryHandler(srv, ctx, dec, interceptor, SalesAdministrationService_PostCustomerReceipt_FullMethodName, func(s SalesAdministrationServiceServer, c context.Context, r *IDRequest) (*CustomerReceipt, error) {
		return s.PostCustomerReceipt(c, r)
	})
}

var SalesAdministrationService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "sales_administration.SalesAdministrationService",
	HandlerType: (*SalesAdministrationServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{MethodName: "CreateSalesInvoice", Handler: _SalesAdministrationService_CreateSalesInvoice_Handler},
		{MethodName: "GetSalesInvoiceByID", Handler: _SalesAdministrationService_GetSalesInvoiceByID_Handler},
		{MethodName: "ListSalesInvoices", Handler: _SalesAdministrationService_ListSalesInvoices_Handler},
		{MethodName: "IssueSalesInvoice", Handler: _SalesAdministrationService_IssueSalesInvoice_Handler},
		{MethodName: "GetSalesInvoiceOutstandingBalance", Handler: _SalesAdministrationService_GetSalesInvoiceOutstandingBalance_Handler},
		{MethodName: "CreateRemittance", Handler: _SalesAdministrationService_CreateRemittance_Handler},
		{MethodName: "GetRemittanceByID", Handler: _SalesAdministrationService_GetRemittanceByID_Handler},
		{MethodName: "ListRemittances", Handler: _SalesAdministrationService_ListRemittances_Handler},
		{MethodName: "ConfirmRemittance", Handler: _SalesAdministrationService_ConfirmRemittance_Handler},
		{MethodName: "CreateCustomerReceipt", Handler: _SalesAdministrationService_CreateCustomerReceipt_Handler},
		{MethodName: "GetCustomerReceiptByID", Handler: _SalesAdministrationService_GetCustomerReceiptByID_Handler},
		{MethodName: "ListCustomerReceipts", Handler: _SalesAdministrationService_ListCustomerReceipts_Handler},
		{MethodName: "PostCustomerReceipt", Handler: _SalesAdministrationService_PostCustomerReceipt_Handler},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "sales_administration.proto",
}
