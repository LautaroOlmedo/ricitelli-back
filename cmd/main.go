package main

import (
	"log"
	"net"
	applicationpb "ricitelli-back/cmd/http/gen/application-service"
	"ricitelli-back/cmd/http/server"
	"ricitelli-back/config"
	repository "ricitelli-back/internal/infraestructure/in-memory"
	application_service "ricitelli-back/internal/service/application-service"
	"ricitelli-back/internal/service/product"
	production_order "ricitelli-back/internal/service/production-order"
	sale_order "ricitelli-back/internal/service/sale-order"

	"google.golang.org/grpc"
)

// TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>
func main() {
	cfg := config.LoadConfig()

	// repository layer
	memoryRepo := repository.NewInMemoryRepository()

	// service layer
	productService := product.NewProductService(&memoryRepo)
	productionOrderService := production_order.NewProductionOrderService(&memoryRepo)
	saleOrderService := sale_order.NewSaleOrderService(&memoryRepo)

	applicationService := application_service.NewApplicationService(productService, productionOrderService, saleOrderService)

	grpcServer := grpc.NewServer()
	svc := server.NewServer(applicationService)
	applicationpb.RegisterApplicationServiceServer(grpcServer, svc)

	lis, socketErr := net.Listen("tcp", cfg.Port)
	if socketErr != nil {
		log.Fatalf("failed to listen: %v", socketErr)
	}

	log.Printf("gRPC server listening on %s\n", cfg.Port)
	if serveErr := grpcServer.Serve(lis); serveErr != nil {
		log.Fatalf("failed to serve: %v", serveErr)
	}

}
