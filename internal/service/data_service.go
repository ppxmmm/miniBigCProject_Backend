package service

import (
	"context"
	"fmt"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
)

// DataService coordinates read-only raw data use cases.
type DataService interface {
	ListStores(ctx context.Context) ([]model.Store, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	ListPaymentMethods(ctx context.Context) ([]model.PaymentMethod, error)
	ListProducts(ctx context.Context) ([]model.ProductEntity, error)
	ListHourlySales(ctx context.Context) ([]model.SalesHourly, error)
	ListDailySales(ctx context.Context) ([]model.SalesDaily, error)
	ListMonthlySales(ctx context.Context) ([]model.SalesMonthly, error)
	ListCategorySales(ctx context.Context) ([]model.CategorySalesEntity, error)
	ListPaymentMix(ctx context.Context) ([]model.PaymentMixEntity, error)
	ListTopProducts(ctx context.Context) ([]model.TopProductEntity, error)
	ListInventoryItems(ctx context.Context) ([]model.InventoryItemEntity, error)
	ListExpiringInventory(ctx context.Context) ([]model.ExpiringInventoryEntity, error)
	ListLowStockAlerts(ctx context.Context) ([]model.LowStockAlertEntity, error)
	ListDeliveries(ctx context.Context) ([]model.Delivery, error)
	ListSuggestions(ctx context.Context) ([]model.Suggestion, error)
}

type dataService struct {
	repository repo.DataRepository
}

// NewDataService creates a DataService.
func NewDataService(repository repo.DataRepository) DataService {
	return &dataService{repository: repository}
}

func (service *dataService) ListStores(ctx context.Context) ([]model.Store, error) {
	rows, err := service.repository.ListStores(ctx)
	if err != nil {
		return nil, fmt.Errorf("list stores: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := service.repository.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListPaymentMethods(ctx context.Context) ([]model.PaymentMethod, error) {
	rows, err := service.repository.ListPaymentMethods(ctx)
	if err != nil {
		return nil, fmt.Errorf("list payment methods: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListProducts(ctx context.Context) ([]model.ProductEntity, error) {
	rows, err := service.repository.ListProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListHourlySales(ctx context.Context) ([]model.SalesHourly, error) {
	rows, err := service.repository.ListHourlySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("list hourly sales: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListDailySales(ctx context.Context) ([]model.SalesDaily, error) {
	rows, err := service.repository.ListDailySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("list daily sales: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListMonthlySales(ctx context.Context) ([]model.SalesMonthly, error) {
	rows, err := service.repository.ListMonthlySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("list monthly sales: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListCategorySales(ctx context.Context) ([]model.CategorySalesEntity, error) {
	rows, err := service.repository.ListCategorySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("list category sales: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListPaymentMix(ctx context.Context) ([]model.PaymentMixEntity, error) {
	rows, err := service.repository.ListPaymentMix(ctx)
	if err != nil {
		return nil, fmt.Errorf("list payment mix: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListTopProducts(ctx context.Context) ([]model.TopProductEntity, error) {
	rows, err := service.repository.ListTopProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list top products: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListInventoryItems(ctx context.Context) ([]model.InventoryItemEntity, error) {
	rows, err := service.repository.ListInventoryItems(ctx)
	if err != nil {
		return nil, fmt.Errorf("list inventory items: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListExpiringInventory(ctx context.Context) ([]model.ExpiringInventoryEntity, error) {
	rows, err := service.repository.ListExpiringInventory(ctx)
	if err != nil {
		return nil, fmt.Errorf("list expiring inventory: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListLowStockAlerts(ctx context.Context) ([]model.LowStockAlertEntity, error) {
	rows, err := service.repository.ListLowStockAlerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("list low stock alerts: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListDeliveries(ctx context.Context) ([]model.Delivery, error) {
	rows, err := service.repository.ListDeliveries(ctx)
	if err != nil {
		return nil, fmt.Errorf("list deliveries: %w", err)
	}

	return rows, nil
}

func (service *dataService) ListSuggestions(ctx context.Context) ([]model.Suggestion, error) {
	rows, err := service.repository.ListSuggestions(ctx)
	if err != nil {
		return nil, fmt.Errorf("list suggestions: %w", err)
	}

	return rows, nil
}
