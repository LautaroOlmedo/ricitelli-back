package main

import (
	"log"
	"net"

	applicationpb "ricitelli-back/cmd/http/gen/application_service"
	drysupplypb "ricitelli-back/cmd/http/gen/dry_supply"
	inventorypb "ricitelli-back/cmd/http/gen/inventory"
	productpb "ricitelli-back/cmd/http/gen/product"
	productionorderpb "ricitelli-back/cmd/http/gen/production_order"
	saleorderpb "ricitelli-back/cmd/http/gen/sale_order"
	"ricitelli-back/cmd/http/server"
	"ricitelli-back/config"
	repository "ricitelli-back/internal/infraestructure/in-memory"
	application_service "ricitelli-back/internal/service/application-service"
	customer_svc "ricitelli-back/internal/service/customer"
	dry_supply_svc "ricitelli-back/internal/service/dry-supply"
	inventory_svc "ricitelli-back/internal/service/inventory"
	"ricitelli-back/internal/service/product"
	product_inventory "ricitelli-back/internal/service/product-inventory"
	production_order "ricitelli-back/internal/service/production-order"
	sale_order "ricitelli-back/internal/service/sale-order"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadConfig()

	// Repository layer (shared in-memory store)
	memoryRepo := repository.NewInMemoryRepository()

	// Domain services
	productService := product.NewProductService(memoryRepo)
	productInventoryService := product_inventory.NewProductInventoryService(memoryRepo)
	productionOrderService := production_order.NewProductionOrderService(memoryRepo)
	saleOrderService := sale_order.NewSaleOrderService(memoryRepo)
	drySupplyService := dry_supply_svc.NewDrySupplyService(memoryRepo)
	customerService := customer_svc.NewCustomerService(memoryRepo, memoryRepo)

	// Application (orchestration) service
	applicationService := application_service.NewApplicationService(
		customerService,
		saleOrderService,
		productionOrderService,
		productService,
		productInventoryService,
		drySupplyService,
	)

	// Inventory service (cross-domain tricapa reads)
	inventoryService := inventory_svc.NewInventoryService(memoryRepo, memoryRepo, memoryRepo)

	// gRPC server
	grpcServer := grpc.NewServer()
	svc := server.NewServer(applicationService, drySupplyService, inventoryService)

	productpb.RegisterProductServiceServer(grpcServer, svc)
	saleorderpb.RegisterSaleOrderServiceServer(grpcServer, svc)
	productionorderpb.RegisterProductionOrderServiceServer(grpcServer, svc)
	applicationpb.RegisterApplicationServiceServer(grpcServer, svc)
	drysupplypb.RegisterDrySupplyServiceServer(grpcServer, svc)
	inventorypb.RegisterInventoryServiceServer(grpcServer, svc)

	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	log.Printf("gRPC server listening on %s\n", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
