package server

import (
	"context"

	authpb "ricitelli-back/cmd/http/gen/auth"
	"ricitelli-back/internal/auth"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	secret    string
	adminUser string
	adminPass string
}

func NewAuthServer(secret, adminUser, adminPass string) *AuthServer {
	return &AuthServer{secret: secret, adminUser: adminUser, adminPass: adminPass}
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	if req.Username == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "username and password are required")
	}
	if req.Username != s.adminUser || req.Password != s.adminPass {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}
	token, err := auth.GenerateToken(req.Username, "admin", s.secret)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not generate token")
	}
	return &authpb.LoginResponse{
		Token:     token,
		ExpiresAt: "24h",
	}, nil
}
