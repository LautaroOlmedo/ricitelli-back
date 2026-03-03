package server

import (
	"context"
	applicationpb "ricitelli-back/cmd/http/gen/application-service"
	productpb "ricitelli-back/cmd/http/gen/product"
	productionorderpb "ricitelli-back/cmd/http/gen/production_order"
	saleorderpb "ricitelli-back/cmd/http/gen/sale_order"
	application_service "ricitelli-back/internal/service/application-service"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	productpb.UnimplementedProductServiceServer
	saleorderpb.UnimplementedSaleOrderServiceServer
	productionorderpb.UnimplementedProductionOrderServiceServer
	applicationpb.UnimplementedApplicationServiceServer

	ApplicationService application_service.Service
}

func NewServer(applicationService application_service.Service) *Server {

	return &Server{
		ApplicationService: applicationService,
	}
}

func (s *Server) CreateOrder(
	ctx context.Context,
	request *applicationpb.CreateOrderRequest,
) (*applicationpb.CreateOrderResponse, error) {
	if request.CustomerId == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid param")
	}

	items := make([]valueObject.SaleOrderItem, 0, len(request.Items))

	for _, item := range request.Items {
		items = append(items, valueObject.SaleOrderItem{
			ProductID: item.ProductId,
			Quantity:  item.Quantity,
		})
	}

	err := s.ApplicationService.CreateOrder(
		ctx,
		request.CustomerId,
		items,
	)

	if err != nil {
		return nil, err
	}

	return &applicationpb.CreateOrderResponse{}, nil
}

func (s *Server) GetSaleOrderByID(
	ctx context.Context,
	request *saleorderpb.GetSaleOrderByIDRequest,
) (*saleorderpb.SaleOrder, error) {

	if request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	order, err := s.ApplicationService.SaleOrderService.GetSaleOrderByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, status.Error(codes.NotFound, "sale order not found")
	}

	items := make([]*saleorderpb.SaleOrderItem, 0, len(order.GetItems()))
	for _, item := range order.GetItems() {
		items = append(items, &saleorderpb.SaleOrderItem{
			ProductId: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	return &saleorderpb.SaleOrder{
		Id:         order.GetID(),
		CustomerId: order.GetCustomerID(),
		Items:      items,
		CreatedAt:  order.GetCreatedAt(),
	}, nil
}

func (s *Server) GetSaleOrders(
	ctx context.Context,
	_ *empty.Empty,
) (*saleorderpb.GetSaleOrdersResponse, error) {

	orders, err := s.ApplicationService.SaleOrderService.GetSaleOrders(ctx)
	if err != nil {
		return nil, err
	}

	responseOrders := make([]*saleorderpb.SaleOrder, 0, len(orders))

	for _, order := range orders {

		items := make([]*saleorderpb.SaleOrderItem, 0, len(order.GetItems()))
		for _, item := range order.GetItems() {
			items = append(items, &saleorderpb.SaleOrderItem{
				ProductId: item.ProductID,
				Quantity:  item.Quantity,
			})
		}

		responseOrders = append(responseOrders, &saleorderpb.SaleOrder{
			Id:         order.GetID(),
			CustomerId: order.GetCustomerID(),
			Items:      items,
			CreatedAt:  order.GetCreatedAt(), // si es string
		})
	}

	return &saleorderpb.GetSaleOrdersResponse{
		SaleOrders: responseOrders,
	}, nil
}

func (s *Server) GetProductionOrderByID(
	ctx context.Context,
	request *productionorderpb.GetProductionOrderByIDRequest,
) (*productionorderpb.ProductionOrder, error) {

	if request.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	order, err := s.ApplicationService.
		ProductionOrderService.
		GetProductionOrderByID(ctx, request.Id)
	if err != nil {
		return nil, err
	}

	if order == nil {
		return nil, status.Error(codes.NotFound, "production order not found")
	}

	// 🔹 Map items
	protoItems := make([]*productionorderpb.ProductionItem, 0, len(order.GetItems()))

	for _, item := range order.GetItems() {

		// 🔹 Map material requirements
		requirements := make([]*productionorderpb.MaterialRequirement, 0, len(item.Requirements))

		for _, req := range item.Requirements {
			requirements = append(requirements, &productionorderpb.MaterialRequirement{
				DrySupplyId: req.DrySupplyID,
				Quantity:    req.Quantity,
			})
		}

		protoItems = append(protoItems, &productionorderpb.ProductionItem{
			ProductId:    item.ProductID,
			Quantity:     item.Quantity,
			Requirements: requirements,
		})
	}

	return &productionorderpb.ProductionOrder{
		Id:          order.GetID(),
		SaleOrderId: order.GetSalesOrderID(),
		Items:       protoItems,
		Status:      order.GetStatus(),
		CreatedAt:   order.GetCreatedAt(),
	}, nil
}

func (s *Server) GetProductionOrders(
	ctx context.Context,
	_ *empty.Empty,
) (*productionorderpb.GetProductionOrdersResponse, error) {

	orders, err := s.ApplicationService.
		ProductionOrderService.
		GetProductionOrders(ctx)
	if err != nil {
		return nil, err
	}

	protoOrders := make([]*productionorderpb.ProductionOrder, 0, len(orders))

	for _, order := range orders {

		// 🔹 Map items
		protoItems := make([]*productionorderpb.ProductionItem, 0, len(order.GetItems()))

		for _, item := range order.GetItems() {

			// 🔹 Map requirements
			requirements := make([]*productionorderpb.MaterialRequirement, 0, len(item.Requirements))

			for _, req := range item.Requirements {
				requirements = append(requirements, &productionorderpb.MaterialRequirement{
					DrySupplyId: req.DrySupplyID,
					Quantity:    req.Quantity,
				})
			}

			protoItems = append(protoItems, &productionorderpb.ProductionItem{
				ProductId:    item.ProductID,
				Quantity:     item.Quantity,
				Requirements: requirements,
			})
		}

		protoOrders = append(protoOrders, &productionorderpb.ProductionOrder{
			Id:          order.GetID(),
			SaleOrderId: order.GetSalesOrderID(),
			Items:       protoItems,
			Status:      order.GetStatus(), // necesita getter
			CreatedAt:   order.GetCreatedAt(),
		})
	}

	return &productionorderpb.GetProductionOrdersResponse{
		ProductionOrders: protoOrders,
	}, nil
}
