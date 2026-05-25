package service

import (
	"context"
	"fmt"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
)

// DashboardService coordinates read-only dashboard use cases.
type DashboardService interface {
	GetDefaultDashboard(ctx context.Context) (model.DashboardData, error)
	GetDashboardByStoreID(ctx context.Context, storeID int64) (model.DashboardData, error)
}

type dashboardService struct {
	repository repo.DashboardRepository
}

// NewDashboardService creates a DashboardService.
func NewDashboardService(repository repo.DashboardRepository) DashboardService {
	return &dashboardService{repository: repository}
}

func (service *dashboardService) GetDefaultDashboard(ctx context.Context) (model.DashboardData, error) {
	storeID, err := service.repository.GetDefaultStoreID(ctx)
	if err != nil {
		return model.DashboardData{}, fmt.Errorf("get default store id: %w", err)
	}

	return service.GetDashboardByStoreID(ctx, storeID)
}

func (service *dashboardService) GetDashboardByStoreID(ctx context.Context, storeID int64) (model.DashboardData, error) {
	data, err := service.repository.GetDashboard(ctx, storeID)
	if err != nil {
		return model.DashboardData{}, fmt.Errorf("get dashboard by store id %d: %w", storeID, err)
	}

	return data, nil
}
