package server

import (
	"context"
	applicationpb "ricitelli-back/cmd/http/gen/application-service"
	saleorderpb "ricitelli-back/cmd/http/gen/sale_order"
	application_service "ricitelli-back/internal/service/application-service"
	valueObject "ricitelli-back/internal/value-object"

	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	applicationpb.UnimplementedApplicationServiceServer
	saleorderpb.UnimplementedSaleOrderServiceServer
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
