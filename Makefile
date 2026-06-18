.PHONY: generate run build

PROTOC  = protoc
MODULE  = ricitelli-back

# Include dirs: all proto directories + system includes (google/protobuf/*)
INCLUDES = \
  -I cmd/http/product \
  -I cmd/http/sale-order \
  -I cmd/http/production-order \
  -I cmd/http/application-service \
  -I cmd/http/dry-supply \
  -I cmd/http/inventory \
  -I cmd/http/customer \
  -I cmd/http/vineyard \
  -I cmd/http/auth \
  -I cmd/http/reporting \
  -I cmd/http/sales-administration \
  -I cmd/http/purchasing-administration \
  -I /opt/homebrew/include

# With module=..., protoc strips the module prefix and outputs files at the
# correct relative path (e.g. ricitelli-back/cmd/http/gen/product → cmd/http/gen/product/)
GO_OUT      = --go_out=. --go_opt=module=$(MODULE)
GO_GRPC_OUT = --go-grpc_out=. --go-grpc_opt=module=$(MODULE)

generate:
	mkdir -p cmd/http/gen/product \
	         cmd/http/gen/sale_order \
	         cmd/http/gen/production_order \
	         cmd/http/gen/application-service \
	         cmd/http/gen/dry_supply \
	         cmd/http/gen/inventory \
	         cmd/http/gen/customer \
	         cmd/http/gen/vineyard \
	         cmd/http/gen/auth \
	         cmd/http/gen/reporting \
	         cmd/http/gen/sales_administration \
	         cmd/http/gen/purchasing_administration

	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/product/product.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/sale-order/sale_order.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/production-order/production_order.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/application-service/application-service.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/dry-supply/dry_supply.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/inventory/inventory.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/customer/customer.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/vineyard/vineyard.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/auth/auth.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/reporting/reporting.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/sales-administration/sales_administration.proto
	$(PROTOC) $(INCLUDES) $(GO_OUT) $(GO_GRPC_OUT) cmd/http/purchasing-administration/purchasing_administration.proto

run:
	go run cmd/main.go

build:
	go build -o bin/ricitelli-back cmd/main.go
