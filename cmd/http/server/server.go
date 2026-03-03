package server

import (
	"context"
	"fmt"
	applicationpb "ricitelli-back/cmd/http/gen/application-service"
	application_service "ricitelli-back/internal/service/application-service"
	valueObject "ricitelli-back/internal/value-object"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
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
	fmt.Println("customer: ", request.CustomerId)

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

	fmt.Println("customer: ", request.CustomerId)

	fmt.Println("Sale Order: ", items)

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
