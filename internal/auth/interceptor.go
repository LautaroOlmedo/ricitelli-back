package auth

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// publicMethods lists fully-qualified gRPC method names that bypass auth.
var publicMethods = map[string]bool{
	"/auth.AuthService/Login": true,
}

var adminServicePrefixes = []string{
	"/sales_administration.SalesAdministrationService/",
	"/purchasing_administration.PurchasingAdministrationService/",
}

func requiresAdmin(fullMethod string) bool {
	for _, prefix := range adminServicePrefixes {
		if strings.HasPrefix(fullMethod, prefix) {
			return true
		}
	}
	return false
}

// NewUnaryInterceptor returns a gRPC unary interceptor that validates JWT tokens.
func NewUnaryInterceptor(secret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}
		tokenStr := strings.TrimPrefix(authHeader[0], "Bearer ")
		if tokenStr == authHeader[0] {
			return nil, status.Error(codes.Unauthenticated, "authorization header must use Bearer scheme")
		}
		claims, err := ValidateToken(tokenStr, secret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token: "+err.Error())
		}
		if requiresAdmin(info.FullMethod) && claims.Role != "admin" {
			return nil, status.Error(codes.PermissionDenied, "admin role required")
		}
		ctx = context.WithValue(ctx, claimsKey{}, claims)
		return handler(ctx, req)
	}
}

type claimsKey struct{}

// ClaimsFromContext extracts JWT claims from context (set by the interceptor).
func ClaimsFromContext(ctx context.Context) (*Claims, bool) {
	c, ok := ctx.Value(claimsKey{}).(*Claims)
	return c, ok
}
