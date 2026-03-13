package server

import (
	"context"

	applicationpb "ricitelli-back/cmd/http/gen/application_service"
	customerpb "ricitelli-back/cmd/http/gen/customer"
	drysupplypb "ricitelli-back/cmd/http/gen/dry_supply"
	inventorypb "ricitelli-back/cmd/http/gen/inventory"
	productpb "ricitelli-back/cmd/http/gen/product"
	productionorderpb "ricitelli-back/cmd/http/gen/production_order"
	saleorderpb "ricitelli-back/cmd/http/gen/sale_order"
	customer_domain "ricitelli-back/internal/domain/customer"
	dry_supply_domain "ricitelli-back/internal/domain/dry-supply"
	production_order_domain "ricitelli-back/internal/domain/production-order"
	sale_order_domain "ricitelli-back/internal/domain/sale-order"
	application_service "ricitelli-back/internal/service/application-service"
	customer_svc "ricitelli-back/internal/service/customer"
	dry_supply_svc "ricitelli-back/internal/service/dry-supply"
	inventory_svc "ricitelli-back/internal/service/inventory"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	productpb.UnimplementedProductServiceServer
	saleorderpb.UnimplementedSaleOrderServiceServer
	productionorderpb.UnimplementedProductionOrderServiceServer
	applicationpb.UnimplementedApplicationServiceServer
	drysupplypb.UnimplementedDrySupplyServiceServer
	inventorypb.UnimplementedInventoryServiceServer
	customerpb.UnimplementedCustomerServiceServer

	AppService       application_service.Service
	DrySupplyService *dry_supply_svc.Service
	InventoryService *inventory_svc.Service
	CustomerService  *customer_svc.Service
}

func NewServer(
	appService application_service.Service,
	drySupplyService *dry_supply_svc.Service,
	inventoryService *inventory_svc.Service,
	customerService *customer_svc.Service,
) *Server {
	return &Server{
		AppService:       appService,
		DrySupplyService: drySupplyService,
		InventoryService: inventoryService,
		CustomerService:  customerService,
	}
}

// ===== ApplicationService =====

func (s *Server) CreateOrder(ctx context.Context, req *applicationpb.CreateOrderRequest) (*applicationpb.CreateOrderResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	items := make([]application_service.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, application_service.OrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	if err := s.AppService.CreateOrder(ctx, application_service.CreateOrderParams{
		CustomerID:         req.CustomerId,
		Items:              items,
		Currency:           sale_order_domain.Currency(req.Currency),
		Market:             sale_order_domain.Market(req.Market),
		DestinationCountry: req.DestinationCountry,
		SaleType:           sale_order_domain.SaleType(req.SaleType),
	}); err != nil {
		return nil, err
	}
	return &applicationpb.CreateOrderResponse{Message: "order created"}, nil
}

// ===== SaleOrderService =====

func (s *Server) CreateSaleOrder(ctx context.Context, req *saleorderpb.CreateSaleOrderRequest) (*saleorderpb.SaleOrder, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	items := make([]valueObject.SaleOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, valueObject.SaleOrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	order, err := s.AppService.SaleOrderService.CreateSaleOrder(ctx, sale_order_domain.NewSaleOrderParams{
		CustomerID:         req.CustomerId,
		Items:              items,
		Currency:           sale_order_domain.Currency(req.Currency),
		Market:             sale_order_domain.Market(req.Market),
		DestinationCountry: req.DestinationCountry,
		SaleType:           sale_order_domain.SaleType(req.SaleType),
	})
	if err != nil {
		return nil, err
	}
	return toProtoSaleOrder(&order), nil
}

func (s *Server) GetSaleOrderByID(ctx context.Context, req *saleorderpb.GetSaleOrderByIDRequest) (*saleorderpb.SaleOrder, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	order, err := s.AppService.SaleOrderService.GetSaleOrderByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "sale order not found")
	}
	return toProtoSaleOrder(order), nil
}

func (s *Server) GetSaleOrders(ctx context.Context, _ *emptypb.Empty) (*saleorderpb.GetSaleOrdersResponse, error) {
	orders, err := s.AppService.SaleOrderService.GetSaleOrders(ctx)
	if err != nil {
		return nil, err
	}
	protoOrders := make([]*saleorderpb.SaleOrder, 0, len(orders))
	for i := range orders {
		protoOrders = append(protoOrders, toProtoSaleOrder(&orders[i]))
	}
	return &saleorderpb.GetSaleOrdersResponse{SaleOrders: protoOrders}, nil
}

func (s *Server) UpdateSaleOrderStatus(ctx context.Context, req *saleorderpb.UpdateSaleOrderStatusRequest) (*saleorderpb.SaleOrder, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	order, err := s.AppService.SaleOrderService.UpdateSaleOrderStatus(ctx, req.Id, sale_order_domain.Status(req.Status))
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoSaleOrder(order), nil
}

func toProtoSaleOrder(o *sale_order_domain.SaleOrder) *saleorderpb.SaleOrder {
	items := make([]*saleorderpb.SaleOrderItem, 0, len(o.GetItems()))
	for _, item := range o.GetItems() {
		items = append(items, &saleorderpb.SaleOrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	return &saleorderpb.SaleOrder{
		Id:                 o.GetID(),
		CustomerId:         o.GetCustomerID(),
		Status:             string(o.GetStatus()),
		Items:              items,
		CreatedAt:          o.GetCreatedAt(),
		Currency:           string(o.GetCurrency()),
		Market:             string(o.GetMarket()),
		DestinationCountry: o.GetDestinationCountry(),
		SaleType:           string(o.GetSaleType()),
	}
}

// ===== ProductionOrderService =====

func (s *Server) GetProductionOrderByID(ctx context.Context, req *productionorderpb.GetProductionOrderByIDRequest) (*productionorderpb.ProductionOrder, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	order, err := s.AppService.ProductionOrderService.GetProductionOrderByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "production order not found")
	}
	return buildProtoProductionOrder(order), nil
}

func (s *Server) GetProductionOrders(ctx context.Context, _ *emptypb.Empty) (*productionorderpb.GetProductionOrdersResponse, error) {
	orders, err := s.AppService.ProductionOrderService.GetProductionOrders(ctx)
	if err != nil {
		return nil, err
	}
	protoOrders := make([]*productionorderpb.ProductionOrder, 0, len(orders))
	for i := range orders {
		protoOrders = append(protoOrders, buildProtoProductionOrder(&orders[i]))
	}
	return &productionorderpb.GetProductionOrdersResponse{ProductionOrders: protoOrders}, nil
}

func buildProtoProductionOrder(order *production_order_domain.ProductionOrder) *productionorderpb.ProductionOrder {
	protoItems := make([]*productionorderpb.ProductionItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		reqs := make([]*productionorderpb.MaterialRequirement, 0, len(item.Requirements))
		for _, r := range item.Requirements {
			reqs = append(reqs, &productionorderpb.MaterialRequirement{
				DrySupplyId: r.DrySupplyID,
				Quantity:    r.Quantity,
			})
		}
		protoItems = append(protoItems, &productionorderpb.ProductionItem{
			ProductId:    item.ProductID,
			Quantity:     item.Quantity,
			Requirements: reqs,
		})
	}
	return &productionorderpb.ProductionOrder{
		Id:          order.GetID(),
		SaleOrderId: order.GetSalesOrderID(),
		Status:      string(order.GetStatus()),
		CreatedAt:   order.GetCreatedAt(),
		Items:       protoItems,
	}
}

// ===== ProductService =====

func (s *Server) GetProductByID(ctx context.Context, req *productpb.GetProductByIDRequest) (*productpb.Product, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	p, err := s.AppService.ProductService.GetProductByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "product not found")
	}
	return toProtoProduct(p), nil
}

func (s *Server) GetProducts(ctx context.Context, _ *emptypb.Empty) (*productpb.GetProductsResponse, error) {
	products, err := s.AppService.ProductService.GetProducts(ctx)
	if err != nil {
		return nil, err
	}
	protoProducts := make([]*productpb.Product, 0, len(products))
	for i := range products {
		protoProducts = append(protoProducts, toProtoProduct(&products[i]))
	}
	return &productpb.GetProductsResponse{Products: protoProducts}, nil
}

func (s *Server) CreateProduct(ctx context.Context, req *productpb.CreateProductRequest) (*emptypb.Empty, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	bods := make([]valueObject.BillOfDrySupply, 0, len(req.Bods))
	for _, b := range req.Bods {
		bods = append(bods, valueObject.BillOfDrySupply{
			DrySupplyID:     b.DrySupplyId,
			QuantityPerUnit: b.QuantityPerUnit,
		})
	}
	if err := s.AppService.ProductService.CreateProduct(ctx, req.Name, bods); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toProtoProduct(p interface {
	GetID() string
	GetName() string
	GetBODS() []valueObject.BillOfDrySupply
}) *productpb.Product {
	bods := make([]*productpb.BillOfDrySupply, 0, len(p.GetBODS()))
	for _, b := range p.GetBODS() {
		bods = append(bods, &productpb.BillOfDrySupply{
			DrySupplyId:     b.DrySupplyID,
			QuantityPerUnit: b.QuantityPerUnit,
		})
	}
	return &productpb.Product{Id: p.GetID(), Name: p.GetName(), Bods: bods}
}

// ===== DrySupplyService =====

func (s *Server) CreateDrySupply(ctx context.Context, req *drysupplypb.CreateDrySupplyRequest) (*drysupplypb.DrySupply, error) {
	if req.Code == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "code and name are required")
	}
	ds, err := s.DrySupplyService.CreateDrySupply(ctx, req.Code, req.Name, dry_supply_domain.Category(req.Category), req.Unit)
	if err != nil {
		return nil, err
	}
	return toProtoDrySupply(ds), nil
}

func (s *Server) GetDrySupplyByID(ctx context.Context, req *drysupplypb.GetDrySupplyByIDRequest) (*drysupplypb.DrySupply, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	ds, err := s.DrySupplyService.GetDrySupplyByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "dry supply not found")
	}
	return toProtoDrySupply(ds), nil
}

func (s *Server) GetDrySupplies(ctx context.Context, _ *emptypb.Empty) (*drysupplypb.GetDrySuppliesResponse, error) {
	supplies, err := s.DrySupplyService.GetDrySupplies(ctx)
	if err != nil {
		return nil, err
	}
	proto := make([]*drysupplypb.DrySupply, 0, len(supplies))
	for i := range supplies {
		proto = append(proto, toProtoDrySupply(&supplies[i]))
	}
	return &drysupplypb.GetDrySuppliesResponse{DrySupplies: proto}, nil
}

func (s *Server) AddStock(ctx context.Context, req *drysupplypb.AddStockRequest) (*emptypb.Empty, error) {
	if err := s.DrySupplyService.AddStock(ctx, req.DrySupplyId, req.Quantity, req.Reference); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetStockTricapa(ctx context.Context, req *drysupplypb.GetStockTricapaRequest) (*drysupplypb.StockTricapa, error) {
	t, err := s.DrySupplyService.GetStockTricapa(ctx, req.DrySupplyId)
	if err != nil {
		return nil, err
	}
	return &drysupplypb.StockTricapa{
		DrySupplyId:    t.DrySupplyID,
		Code:           t.DrySupplyCode,
		Name:           t.DrySupplyName,
		PhysicalStock:  t.Physical,
		CommittedStock: t.Committed,
		AvailableStock: t.Available,
	}, nil
}

func (s *Server) CommitStock(ctx context.Context, req *drysupplypb.CommitStockRequest) (*emptypb.Empty, error) {
	if err := s.DrySupplyService.CommitStock(ctx, req.DrySupplyId, req.Quantity, req.ProductionOrderId); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ReleaseStock(ctx context.Context, req *drysupplypb.ReleaseStockRequest) (*emptypb.Empty, error) {
	if err := s.DrySupplyService.ReleaseStock(ctx, req.DrySupplyId, req.Quantity, req.Reference); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) ConsumeStock(ctx context.Context, req *drysupplypb.ConsumeStockRequest) (*emptypb.Empty, error) {
	if err := s.DrySupplyService.ConsumeStock(ctx, req.DrySupplyId, req.Quantity, req.Reference); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func toProtoDrySupply(ds interface {
	GetID() string
	GetCode() string
	GetName() string
	GetCategory() dry_supply_domain.Category
	GetUnit() string
}) *drysupplypb.DrySupply {
	return &drysupplypb.DrySupply{
		Id:       ds.GetID(),
		Code:     ds.GetCode(),
		Name:     ds.GetName(),
		Category: string(ds.GetCategory()),
		Unit:     ds.GetUnit(),
	}
}

// ===== InventoryService =====

func (s *Server) GetInventoryReport(ctx context.Context, _ *emptypb.Empty) (*inventorypb.InventoryReport, error) {
	report, err := s.InventoryService.GetInventoryReport(ctx)
	if err != nil {
		return nil, err
	}
	return toProtoInventoryReport(report), nil
}

func (s *Server) GetLowStockAlerts(ctx context.Context, _ *emptypb.Empty) (*inventorypb.InventoryReport, error) {
	alerts, err := s.InventoryService.GetLowStockAlerts(ctx)
	if err != nil {
		return nil, err
	}
	protoAlerts := make([]*inventorypb.DrySupplyAlert, 0, len(alerts))
	for _, a := range alerts {
		protoAlerts = append(protoAlerts, &inventorypb.DrySupplyAlert{
			DrySupplyId: a.DrySupplyID,
			Code:        a.Code,
			Name:        a.Name,
			Physical:    a.Physical,
			Committed:   a.Committed,
			Available:   a.Available,
			IsLow:       a.IsLow,
		})
	}
	return &inventorypb.InventoryReport{DrySupplyAlerts: protoAlerts}, nil
}

func (s *Server) ConvertSVtoPT(ctx context.Context, req *inventorypb.ConvertSVtoPTRequest) (*emptypb.Empty, error) {
	if req.ProductId == "" || req.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and quantity are required")
	}
	if err := s.InventoryService.ConvertSVtoPT(ctx, req.ProductId, req.Quantity, req.LotNumber); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) AddUndressedStock(ctx context.Context, req *inventorypb.AddUndressedStockRequest) (*emptypb.Empty, error) {
	if req.ProductId == "" || req.Quantity == 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id and quantity are required")
	}
	if err := s.InventoryService.AddUndressedStock(ctx, req.ProductId, req.Quantity, req.Reference); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetProductTricapa(ctx context.Context, req *inventorypb.GetProductTricapaRequest) (*inventorypb.ProductTricapa, error) {
	t, err := s.InventoryService.GetProductTricapa(ctx, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &inventorypb.ProductTricapa{
		ProductId:        t.ProductID,
		ProductName:      t.ProductName,
		Sku:              t.SKU,
		UndressedStock:   t.UndressedStock,
		DressedPhysical:  t.DressedPhysical,
		DressedCommitted: t.DressedCommitted,
		DressedAvailable: t.DressedAvailable,
	}, nil
}

func toProtoInventoryReport(r *inventory_svc.InventoryReport) *inventorypb.InventoryReport {
	protoProducts := make([]*inventorypb.ProductTricapa, 0, len(r.Products))
	for _, p := range r.Products {
		protoProducts = append(protoProducts, &inventorypb.ProductTricapa{
			ProductId:        p.ProductID,
			ProductName:      p.ProductName,
			Sku:              p.SKU,
			UndressedStock:   p.UndressedStock,
			DressedPhysical:  p.DressedPhysical,
			DressedCommitted: p.DressedCommitted,
			DressedAvailable: p.DressedAvailable,
		})
	}
	protoAlerts := make([]*inventorypb.DrySupplyAlert, 0, len(r.DrySupplyAlerts))
	for _, a := range r.DrySupplyAlerts {
		protoAlerts = append(protoAlerts, &inventorypb.DrySupplyAlert{
			DrySupplyId: a.DrySupplyID,
			Code:        a.Code,
			Name:        a.Name,
			Physical:    a.Physical,
			Committed:   a.Committed,
			Available:   a.Available,
			IsLow:       a.IsLow,
		})
	}
	return &inventorypb.InventoryReport{
		Products:        protoProducts,
		DrySupplyAlerts: protoAlerts,
	}
}

// ===== CustomerService =====

func (s *Server) CreateCustomer(ctx context.Context, req *customerpb.CreateCustomerRequest) (*customerpb.Customer, error) {
	if req.SocialReason == "" {
		return nil, status.Error(codes.InvalidArgument, "social_reason is required")
	}
	c, err := s.CustomerService.CreateCustomer(ctx, customer_domain.NewCustomerParams{
		SocialReason: req.SocialReason,
		MarketType:   customer_domain.MarketType(req.MarketType),
		Group:        customer_domain.Group(req.Group),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoCustomer(&c), nil
}

func (s *Server) GetCustomerByID(ctx context.Context, req *customerpb.GetCustomerByIDRequest) (*customerpb.Customer, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	c, err := s.CustomerService.GetCustomerByID(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "customer not found")
	}
	return toProtoCustomer(c), nil
}

func (s *Server) GetCustomers(ctx context.Context, _ *empty.Empty) (*customerpb.GetCustomersResponse, error) {
	customers, err := s.CustomerService.GetCustomers(ctx)
	if err != nil {
		return nil, err
	}
	proto := make([]*customerpb.Customer, 0, len(customers))
	for i := range customers {
		proto = append(proto, toProtoCustomer(&customers[i]))
	}
	return &customerpb.GetCustomersResponse{Customers: proto}, nil
}

func (s *Server) DeactivateCustomer(ctx context.Context, req *customerpb.DeactivateCustomerRequest) (*customerpb.Customer, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	c, err := s.CustomerService.DeactivateCustomer(ctx, req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoCustomer(c), nil
}

func (s *Server) PlaceOrder(ctx context.Context, req *customerpb.PlaceOrderRequest) (*saleorderpb.SaleOrder, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	items := make([]valueObject.SaleOrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, valueObject.SaleOrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
			UnitPrice: item.UnitPrice,
		})
	}
	order, err := s.CustomerService.PlaceOrder(ctx, customer_svc.PlaceOrderParams{
		CustomerID:         req.CustomerId,
		Items:              items,
		Currency:           sale_order_domain.Currency(req.Currency),
		DestinationCountry: req.DestinationCountry,
		SaleType:           sale_order_domain.SaleType(req.SaleType),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoSaleOrder(&order), nil
}

func (s *Server) GetOrdersByCustomer(ctx context.Context, req *customerpb.GetOrdersByCustomerRequest) (*customerpb.GetOrdersByCustomerResponse, error) {
	if req.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "customer_id is required")
	}
	orders, err := s.CustomerService.GetOrdersByCustomer(ctx, req.CustomerId)
	if err != nil {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	protoOrders := make([]*saleorderpb.SaleOrder, 0, len(orders))
	for i := range orders {
		protoOrders = append(protoOrders, toProtoSaleOrder(&orders[i]))
	}
	return &customerpb.GetOrdersByCustomerResponse{Orders: protoOrders}, nil
}

func toProtoCustomer(c interface {
	GetID() string
	GetSocialReason() string
	GetMarketType() customer_domain.MarketType
	GetGroup() customer_domain.Group
	IsActive() bool
	GetCreatedAt() string
}) *customerpb.Customer {
	return &customerpb.Customer{
		Id:           c.GetID(),
		SocialReason: c.GetSocialReason(),
		MarketType:   string(c.GetMarketType()),
		Group:        string(c.GetGroup()),
		Active:       c.IsActive(),
		CreatedAt:    c.GetCreatedAt(),
	}
}

// ===== New handlers added for BE-03, BE-06 =====

func (s *Server) UpdateCustomer(ctx context.Context, req *customerpb.UpdateCustomerRequest) (*customerpb.Customer, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	c, err := s.CustomerService.UpdateCustomer(ctx, req.Id, customer_domain.UpdateCustomerParams{
		SocialReason: req.SocialReason,
		MarketType:   customer_domain.MarketType(req.MarketType),
		Group:        customer_domain.Group(req.Group),
	})
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return toProtoCustomer(c), nil
}

func (s *Server) UpdateProduct(ctx context.Context, req *productpb.UpdateProductRequest) (*emptypb.Empty, error) {
	if req.Id == "" || req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "id and name are required")
	}
	bods := make([]valueObject.BillOfDrySupply, 0, len(req.Bods))
	for _, b := range req.Bods {
		bods = append(bods, valueObject.BillOfDrySupply{DrySupplyID: b.DrySupplyId, QuantityPerUnit: b.QuantityPerUnit})
	}
	if err := s.AppService.ProductService.UpdateProduct(ctx, req.Id, req.Name, bods); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) UpdateDrySupply(ctx context.Context, req *drysupplypb.UpdateDrySupplyRequest) (*emptypb.Empty, error) {
	if req.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.DrySupplyService.UpdateDrySupply(ctx, req.Id, req.Name, int(req.ReorderPoint)); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Server) GetProductionOrdersBySaleOrder(ctx context.Context, req *productionorderpb.GetProductionOrdersBySaleOrderRequest) (*productionorderpb.GetProductionOrdersResponse, error) {
	if req.SaleOrderId == "" {
		return nil, status.Error(codes.InvalidArgument, "sale_order_id is required")
	}
	orders, err := s.AppService.ProductionOrderService.GetProductionOrdersBySaleOrder(ctx, req.SaleOrderId)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	protoOrders := make([]*productionorderpb.ProductionOrder, 0, len(orders))
	for i := range orders {
		protoOrders = append(protoOrders, buildProtoProductionOrder(&orders[i]))
	}
	return &productionorderpb.GetProductionOrdersResponse{ProductionOrders: protoOrders}, nil
}
