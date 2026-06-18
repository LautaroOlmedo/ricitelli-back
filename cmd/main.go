package main

import (
	"flag"
	"log"
	"net"
	"net/http"

	applicationpb "ricitelli-back/cmd/http/gen/application_service"
	authpb "ricitelli-back/cmd/http/gen/auth"
	customerpb "ricitelli-back/cmd/http/gen/customer"
	drysupplypb "ricitelli-back/cmd/http/gen/dry_supply"
	inventorypb "ricitelli-back/cmd/http/gen/inventory"
	productpb "ricitelli-back/cmd/http/gen/product"
	productionorderpb "ricitelli-back/cmd/http/gen/production_order"
	purchasingadministrationpb "ricitelli-back/cmd/http/gen/purchasing_administration"
	reportingpb "ricitelli-back/cmd/http/gen/reporting"
	saleorderpb "ricitelli-back/cmd/http/gen/sale_order"
	salesadministrationpb "ricitelli-back/cmd/http/gen/sales_administration"
	vineyardpb "ricitelli-back/cmd/http/gen/vineyard"
	"ricitelli-back/cmd/http/server"
	"ricitelli-back/config"
	"ricitelli-back/internal/auth"
	inmemory "ricitelli-back/internal/infraestructure/in-memory"
	pgstore "ricitelli-back/internal/infraestructure/postgres"
	application_service "ricitelli-back/internal/service/application-service"
	customer_svc "ricitelli-back/internal/service/customer"
	dry_supply_svc "ricitelli-back/internal/service/dry-supply"
	inventory_svc "ricitelli-back/internal/service/inventory"
	"ricitelli-back/internal/service/product"
	product_inventory "ricitelli-back/internal/service/product-inventory"
	production_order "ricitelli-back/internal/service/production-order"
	purchasing_administration "ricitelli-back/internal/service/purchasing-administration"
	reporting_svc "ricitelli-back/internal/service/reporting"
	reporting_storage "ricitelli-back/internal/service/reporting/storage"
	sale_order "ricitelli-back/internal/service/sale-order"
	sales_administration "ricitelli-back/internal/service/sales-administration"
	vineyard_svc "ricitelli-back/internal/service/vineyard"

	"google.golang.org/grpc"
)

func main() {
	minimal := flag.Bool("minimal", false, "arranca sin datos de prueba (solo clientes + usuario admin)")
	flag.Parse()

	cfg := config.LoadConfig()

	if cfg.DatabaseURL != "" {
		log.Printf("connecting to PostgreSQL…\n")
		pgRepo, err := pgstore.NewRepository(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("failed to connect to database: %v", err)
		}
		log.Println("PostgreSQL connected and migrations applied")
		// vineyard uses in-memory until postgres storage is implemented
		vineyardRepo := inmemory.NewInMemoryRepository()
		boot(cfg,
			product.NewProductService(pgRepo),
			product_inventory.NewProductInventoryService(pgRepo),
			production_order.NewProductionOrderService(pgRepo),
			sale_order.NewSaleOrderService(pgRepo),
			dry_supply_svc.NewDrySupplyService(pgRepo),
			customer_svc.NewCustomerService(pgRepo, pgRepo),
			vineyard_svc.NewVineyardService(vineyardRepo),
			pgRepo,
			pgRepo,
			pgRepo,
		)
		return
	}

	log.Println("DATABASE_URL not set — using in-memory repository")
	var r *inmemory.InMemoryRepository
	if *minimal {
		r = inmemory.NewInMemoryRepositoryMinimal()
	} else {
		r = inmemory.NewInMemoryRepository()
	}
	boot(cfg,
		product.NewProductService(r),
		product_inventory.NewProductInventoryService(r),
		production_order.NewProductionOrderService(r),
		sale_order.NewSaleOrderService(r),
		dry_supply_svc.NewDrySupplyService(r),
		customer_svc.NewCustomerService(r, r),
		vineyard_svc.NewVineyardService(r),
		r,
		r,
		r,
	)
}

func boot(
	cfg config.Config,
	productSvc *product.ProductService,
	productInvSvc *product_inventory.Service,
	productionOrderSvc *production_order.Service,
	saleOrderSvc *sale_order.Service,
	drySupplySvc *dry_supply_svc.Service,
	customerSvc *customer_svc.Service,
	vineyardSvc *vineyard_svc.Service,
	movementStorage inventory_svc.MovementStorage,
	salesAdministrationStorage sales_administration.Storage,
	purchasingAdministrationStorage purchasing_administration.Storage,
) {
	appSvc := application_service.NewApplicationService(
		customerSvc, saleOrderSvc, productionOrderSvc,
		productSvc, productInvSvc, drySupplySvc,
	)

	inventorySvc := inventory_svc.NewInventoryService(productSvc, productInvSvc, drySupplySvc, movementStorage)

	// Reporting service (reuses existing services — agnostic of repository).
	reportingStore, err := reporting_storage.NewStore(cfg.ReportsOutputDir)
	if err != nil {
		log.Fatalf("failed to init reporting storage: %v", err)
	}
	reportingSvc := reporting_svc.NewService(
		saleOrderSvc,
		productionOrderSvc,
		productSvc,
		drySupplySvc,
		customerSvc,
		inventorySvc,
		reportingStore,
		cfg.ReportsPublicURLBase,
	)
	salesAdministrationSvc := sales_administration.NewService(salesAdministrationStorage)
	purchasingAdministrationSvc := purchasing_administration.NewService(purchasingAdministrationStorage)

	interceptor := auth.NewUnaryInterceptor(cfg.JWTSecret)
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(interceptor))

	svc := server.NewServer(appSvc, drySupplySvc, inventorySvc, customerSvc, vineyardSvc)
	authSvc := server.NewAuthServer(cfg.JWTSecret, cfg.AdminUser, cfg.AdminPass)
	reportingSrv := server.NewReportingServer(reportingSvc)
	salesAdministrationSrv := server.NewSalesAdministrationServer(salesAdministrationSvc)
	purchasingAdministrationSrv := server.NewPurchasingAdministrationServer(purchasingAdministrationSvc)

	productpb.RegisterProductServiceServer(grpcServer, svc)
	saleorderpb.RegisterSaleOrderServiceServer(grpcServer, svc)
	productionorderpb.RegisterProductionOrderServiceServer(grpcServer, svc)
	applicationpb.RegisterApplicationServiceServer(grpcServer, svc)
	drysupplypb.RegisterDrySupplyServiceServer(grpcServer, svc)
	inventorypb.RegisterInventoryServiceServer(grpcServer, svc)
	customerpb.RegisterCustomerServiceServer(grpcServer, svc)
	vineyardpb.RegisterVineyardServiceServer(grpcServer, svc)
	authpb.RegisterAuthServiceServer(grpcServer, authSvc)
	reportingpb.RegisterReportingServiceServer(grpcServer, reportingSrv)
	salesadministrationpb.RegisterSalesAdministrationServiceServer(grpcServer, salesAdministrationSrv)
	purchasingadministrationpb.RegisterPurchasingAdministrationServiceServer(grpcServer, purchasingAdministrationSrv)

	// HTTP server for downloading generated PDFs.
	go func() {
		handler := server.NewDownloadHandler(reportingSvc, cfg.JWTSecret)
		log.Printf("HTTP report download server listening on %s\n", cfg.ReportsHTTPPort)
		if err := http.ListenAndServe(cfg.ReportsHTTPPort, handler); err != nil {
			log.Printf("report download server stopped: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", cfg.Port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Printf("gRPC server listening on %s\n", cfg.Port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
