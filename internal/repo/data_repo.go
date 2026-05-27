package repo

import (
	"context"
	"fmt"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"gorm.io/gorm"
)

// DataRepository reads raw records from all dashboard tables.
type DataRepository interface {
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

type gormDataRepository struct {
	db *gorm.DB
}

// NewDataRepository creates a GORM-backed raw data repository.
func NewDataRepository(db *gorm.DB) DataRepository {
	return &gormDataRepository{db: db}
}

func (repo *gormDataRepository) ListStores(ctx context.Context) ([]model.Store, error) {
	var rows []model.Store
	if err := repo.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query stores: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListCategories(ctx context.Context) ([]model.Category, error) {
	var rows []model.Category
	if err := repo.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query categories: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListPaymentMethods(ctx context.Context) ([]model.PaymentMethod, error) {
	var rows []model.PaymentMethod
	if err := repo.db.WithContext(ctx).Order("id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query payment methods: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListProducts(ctx context.Context) ([]model.ProductEntity, error) {
	var rows []model.ProductEntity
	if err := repo.db.WithContext(ctx).Order("sku").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query products: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListHourlySales(ctx context.Context) ([]model.SalesHourly, error) {
	var rows []model.SalesHourly
	if err := repo.db.WithContext(ctx).Order("store_id, sale_date, hour").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query hourly sales: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListDailySales(ctx context.Context) ([]model.SalesDaily, error) {
	var rows []model.SalesDaily
	if err := repo.db.WithContext(ctx).Order("store_id, sale_date").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query daily sales: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListMonthlySales(ctx context.Context) ([]model.SalesMonthly, error) {
	var rows []model.SalesMonthly
	if err := repo.db.WithContext(ctx).Order("store_id, year, month").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query monthly sales: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListCategorySales(ctx context.Context) ([]model.CategorySalesEntity, error) {
	var rows []model.CategorySalesEntity
	if err := repo.db.WithContext(ctx).Order("store_id, sales_date, category_id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query category sales: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListPaymentMix(ctx context.Context) ([]model.PaymentMixEntity, error) {
	var rows []model.PaymentMixEntity
	if err := repo.db.WithContext(ctx).Order("store_id, sales_date, payment_method_id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query payment mix: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListTopProducts(ctx context.Context) ([]model.TopProductEntity, error) {
	var rows []model.TopProductEntity
	if err := repo.db.WithContext(ctx).Order("store_id, sales_date, sales_value DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query top products: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListInventoryItems(ctx context.Context) ([]model.InventoryItemEntity, error) {
	var rows []model.InventoryItemEntity
	if err := repo.db.WithContext(ctx).Order("store_id, location_code, sku").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query inventory items: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListExpiringInventory(ctx context.Context) ([]model.ExpiringInventoryEntity, error) {
	var rows []model.ExpiringInventoryEntity
	if err := repo.db.WithContext(ctx).Order("store_id, expiry_date, sku").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query expiring inventory: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListLowStockAlerts(ctx context.Context) ([]model.LowStockAlertEntity, error) {
	var rows []model.LowStockAlertEntity
	if err := repo.db.WithContext(ctx).Order("store_id, stock_quantity, sku").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query low stock alerts: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListDeliveries(ctx context.Context) ([]model.Delivery, error) {
	var rows []model.Delivery
	if err := repo.db.WithContext(ctx).Order("store_id, eta_time, id").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query deliveries: %w", err)
	}

	return rows, nil
}

func (repo *gormDataRepository) ListSuggestions(ctx context.Context) ([]model.Suggestion, error) {
	var rows []model.Suggestion
	if err := repo.db.WithContext(ctx).Order("store_id, kind DESC, upside_value DESC").Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("query suggestions: %w", err)
	}

	return rows, nil
}
