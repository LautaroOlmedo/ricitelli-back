package dependencies

import (
	"fmt"
	repository "ricitelli-back/internal/infraestructure/in-memory"
	application_service "ricitelli-back/internal/service/application-service"
	"ricitelli-back/internal/service/product"
	production_order "ricitelli-back/internal/service/production-order"
	sale_order "ricitelli-back/internal/service/sale-order"
	"time"
)

type Dependencies struct {
}

func InitDependencies(cfg config.Config) Dependencies {
	// repository layer
	memoryRepo := repository.NewInMemoryRepository()

	// service layer
	productService := product.NewProductService(&memoryRepo)
	productionOrderService := production_order.NewProductionOrderService(&memoryRepo)
	saleOrderService := sale_order.NewSaleOrderService(&memoryRepo)

	applicationService := application_service.NewApplicationService(productService, productionOrderService, saleOrderService)

	// handler layer
	writerHandler := writer.NewWriteHandler(productsService, ordersService)
	readerHandler := reader.NewReaderHandler(productsService, ordersService, tokenGenerator)

	return Dependencies{
		WriterHandler: *writerHandler,
		ReaderHandler: *readerHandler,
	}

}
