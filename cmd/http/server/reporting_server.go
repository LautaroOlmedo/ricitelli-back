package server

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strings"

	reportingpb "ricitelli-back/cmd/http/gen/reporting"
	"ricitelli-back/internal/auth"
	reporting_svc "ricitelli-back/internal/service/reporting"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ReportingServer implements the gRPC ReportingService.
type ReportingServer struct {
	reportingpb.UnimplementedReportingServiceServer
	Svc *reporting_svc.Service
}

func NewReportingServer(svc *reporting_svc.Service) *ReportingServer {
	return &ReportingServer{Svc: svc}
}

func toFilter(req *reportingpb.ReportFilter) reporting_svc.ReportFilter {
	return reporting_svc.ReportFilter{
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		CustomerID: req.CustomerId,
		Market:     req.Market,
		Currency:   req.Currency,
		ProductID:  req.ProductId,
	}
}

func toProto(r *reporting_svc.Result) *reportingpb.ReportResponse {
	if r == nil {
		return nil
	}
	return &reportingpb.ReportResponse{
		Id:          r.ID,
		Type:        string(r.Type),
		Filename:    r.Filename,
		DownloadUrl: r.DownloadURL,
		GeneratedAt: r.GeneratedAt,
		FileSize:    r.FileSize,
		FromDate:    r.FromDate,
		ToDate:      r.ToDate,
	}
}

func (s *ReportingServer) GenerateSalesReport(ctx context.Context, req *reportingpb.ReportFilter) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateSalesReport(ctx, toFilter(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) GenerateProductionReport(ctx context.Context, req *reportingpb.ReportFilter) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateProductionReport(ctx, toFilter(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) GenerateGeneralReport(ctx context.Context, req *reportingpb.ReportFilter) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateGeneralReport(ctx, toFilter(req))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) GenerateLowStockReport(ctx context.Context, _ *emptypb.Empty) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateLowStockReport(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) GenerateLotTraceabilityReport(ctx context.Context, req *reportingpb.LotRequest) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateLotTraceabilityReport(ctx, req.LotNumber)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) GenerateCustomerReport(ctx context.Context, req *reportingpb.CustomerReportRequest) (*reportingpb.ReportResponse, error) {
	r, err := s.Svc.GenerateCustomerReport(ctx, req.CustomerId, req.FromDate, req.ToDate)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProto(r), nil
}

func (s *ReportingServer) ListReports(ctx context.Context, req *reportingpb.ListReportsRequest) (*reportingpb.ListReportsResponse, error) {
	results, total, err := s.Svc.ListReports(ctx, req.Type, int(req.Page), int(req.PageSize))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := make([]*reportingpb.ReportResponse, len(results))
	for i, r := range results {
		out[i] = toProto(r)
	}
	return &reportingpb.ListReportsResponse{Reports: out, TotalCount: int32(total)}, nil
}

// ---------------------------------------------------------------------------
// HTTP download handler
// ---------------------------------------------------------------------------

// NewDownloadHandler returns an http.Handler that serves PDFs stored by the service.
// URL pattern: /reports/files/{id}. Auth: Bearer token (header or ?token= query).
func NewDownloadHandler(svc *reporting_svc.Service, jwtSecret string) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/reports/files/", func(w http.ResponseWriter, r *http.Request) {
		// Auth
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == r.Header.Get("Authorization") {
			// No Bearer prefix — try query param
			token = r.URL.Query().Get("token")
		}
		if token == "" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if _, err := auth.ValidateToken(token, jwtSecret); err != nil {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// ID from path
		id := strings.TrimPrefix(r.URL.Path, "/reports/files/")
		id = strings.TrimSuffix(id, "/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		m, ok := svc.Storage.Get(id)
		if !ok {
			http.Error(w, "report not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filepath.Base(m.Filename)))
		http.ServeFile(w, r, m.Path)
	})
	return mux
}
