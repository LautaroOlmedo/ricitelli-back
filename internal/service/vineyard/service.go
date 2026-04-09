package vineyard

import (
	"context"

	"ricitelli-back/internal/domain/vineyard"
)

type VineyardStorage interface {
	CreatePlot(ctx context.Context, name, variety string, ha float64, age int32, status string, polygon []vineyard.LatLng) (*vineyard.Plot, error)
	UpdatePlot(ctx context.Context, id, name, variety string, ha float64, age int32, status string, polygon []vineyard.LatLng) (*vineyard.Plot, error)
	DeletePlot(ctx context.Context, id string) error
	GetPlotByID(ctx context.Context, id string) (*vineyard.Plot, error)
	GetPlots(ctx context.Context) ([]vineyard.Plot, error)
}

type Service struct {
	Storage VineyardStorage
}

func NewVineyardService(storage VineyardStorage) *Service {
	return &Service{Storage: storage}
}

func (s *Service) CreatePlot(ctx context.Context, name, variety string, ha float64, age int32, status string, polygon []vineyard.LatLng) (*vineyard.Plot, error) {
	return s.Storage.CreatePlot(ctx, name, variety, ha, age, status, polygon)
}

func (s *Service) UpdatePlot(ctx context.Context, id, name, variety string, ha float64, age int32, status string, polygon []vineyard.LatLng) (*vineyard.Plot, error) {
	return s.Storage.UpdatePlot(ctx, id, name, variety, ha, age, status, polygon)
}

func (s *Service) DeletePlot(ctx context.Context, id string) error {
	return s.Storage.DeletePlot(ctx, id)
}

func (s *Service) GetPlotByID(ctx context.Context, id string) (*vineyard.Plot, error) {
	return s.Storage.GetPlotByID(ctx, id)
}

func (s *Service) GetPlots(ctx context.Context) ([]vineyard.Plot, error) {
	return s.Storage.GetPlots(ctx)
}
