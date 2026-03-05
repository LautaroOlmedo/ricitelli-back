protoc -I=.   -I=../product   -I=../sale-order   -I=../production-order   --go_out=paths=source_relative:../gen/application-service   --go-grpc_out=paths=source_relative:../gen/application-service   application-service.proto


protoc   -I=.   -I=../product   -I=../sale-order   -I=../application-service   --go_out=paths=source_relative:../gen/production_order   --go-grpc_out=paths=source_relative:../gen/production_order   production_order.proto


protoc   -I=.   -I=../product   -I=../production-order   -I=../application-service   --go_out=paths=source_relative:../gen/sale_order   --go-grpc_out=paths=source_relative:../gen/sale_order   sale_order.proto


protoc   -I=.   -I=../sale-order   -I=../production-order   -I=../application-service   --go_out=paths=source_relative:../gen/product   --go-grpc_out=paths=source_relative:../gen/product   product.proto