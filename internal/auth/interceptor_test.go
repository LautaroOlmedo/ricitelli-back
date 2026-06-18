package auth

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestUnaryInterceptorAdministrativeMethodsRequireAdmin(t *testing.T) {
	const secret = "test-secret"
	interceptor := NewUnaryInterceptor(secret)
	handler := func(ctx context.Context, req any) (any, error) { return "ok", nil }

	for _, method := range []string{
		"/sales_administration.SalesAdministrationService/CreateSalesInvoice",
		"/purchasing_administration.PurchasingAdministrationService/CreateSupplier",
	} {
		t.Run(method, func(t *testing.T) {
			token, err := GenerateToken("user-1", "user", secret)
			if err != nil {
				t.Fatal(err)
			}
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

			_, err = interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: method}, handler)
			if status.Code(err) != codes.PermissionDenied {
				t.Fatalf("expected PermissionDenied, got %v", err)
			}
		})
	}
}

func TestUnaryInterceptorAdminCanCallAdministrativeMethods(t *testing.T) {
	const secret = "test-secret"
	token, err := GenerateToken("admin-1", "admin", secret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	_, err = NewUnaryInterceptor(secret)(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/sales_administration.SalesAdministrationService/CreateSalesInvoice"},
		func(ctx context.Context, req any) (any, error) { return "ok", nil },
	)
	if err != nil {
		t.Fatalf("expected admin call to succeed, got %v", err)
	}
}

func TestUnaryInterceptorExistingAuthenticatedMethodsAcceptNonAdmin(t *testing.T) {
	const secret = "test-secret"
	token, err := GenerateToken("user-1", "user", secret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("authorization", "Bearer "+token))

	_, err = NewUnaryInterceptor(secret)(
		ctx,
		nil,
		&grpc.UnaryServerInfo{FullMethod: "/product.ProductService/GetProducts"},
		func(ctx context.Context, req any) (any, error) {
			claims, ok := ClaimsFromContext(ctx)
			if !ok || claims.UserID != "user-1" {
				t.Fatalf("expected claims in handler context")
			}
			return "ok", nil
		},
	)
	if err != nil {
		t.Fatalf("expected existing authenticated call to succeed, got %v", err)
	}
}
