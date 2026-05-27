package repo

import (
	"context"
	"errors"
	"fmt"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"gorm.io/gorm"
)

// ErrNotFound is returned when a requested record does not exist.
var ErrNotFound = errors.New("record not found")

// DashboardRepository reads dashboard data from PostgreSQL.
type DashboardRepository interface {
	GetDashboard(ctx context.Context, storeID int64) (model.DashboardData, error)
	GetDefaultStoreID(ctx context.Context) (int64, error)
}

type gormDashboardRepository struct {
	db *gorm.DB
}

// NewDashboardRepository creates a GORM-backed dashboard repository.
func NewDashboardRepository(db *gorm.DB) DashboardRepository {
	return &gormDashboardRepository{db: db}
}

func (repo *gormDashboardRepository) GetDefaultStoreID(ctx context.Context) (int64, error) {
	var store model.Store
	err := repo.db.WithContext(ctx).Order("id").First(&store).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("query default store id: %w", err)
	}

	return store.ID, nil
}

func (repo *gormDashboardRepository) GetDashboard(ctx context.Context, storeID int64) (model.DashboardData, error) {
	var data model.DashboardData

	if err := repo.db.WithContext(ctx).First(&data.Store, storeID).Error; errors.Is(err, gorm.ErrRecordNotFound) {
		return data, ErrNotFound
	} else if err != nil {
		return data, fmt.Errorf("query store %d: %w", storeID, err)
	}

	loaders := []func(context.Context, int64, *model.DashboardData) error{
		repo.loadHourlySales,
		repo.loadDailySales,
		repo.loadMonthlySales,
		repo.loadCategorySales,
		repo.loadPaymentMix,
		repo.loadTopProducts,
		repo.loadInventoryItems,
		repo.loadExpiringInventory,
		repo.loadLowStockAlerts,
		repo.loadDeliveries,
		repo.loadSuggestions,
	}
	for _, loader := range loaders {
		if err := loader(ctx, storeID, &data); err != nil {
			return data, err
		}
	}

	return data, nil
}

func (repo *gormDashboardRepository) loadHourlySales(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("sales_hourly").
		Where("store_id = ?", storeID).
		Order("sale_date, hour").
		Find(&data.HourlySales).Error
	if err != nil {
		return fmt.Errorf("query hourly sales: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadDailySales(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("sales_daily").
		Where("store_id = ?", storeID).
		Order("sale_date").
		Find(&data.DailySales).Error
	if err != nil {
		return fmt.Errorf("query daily sales: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadMonthlySales(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("sales_monthly").
		Where("store_id = ?", storeID).
		Order("year, month").
		Find(&data.MonthlySales).Error
	if err != nil {
		return fmt.Errorf("query monthly sales: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadCategorySales(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("category_sales cs").
		Select(`
			cs.id, cs.store_id, cs.category_id, c.name_th, c.name_en, c.color,
			cs.sales_date, cs.sales_value, cs.share, cs.trend_percent`).
		Joins("JOIN categories c ON c.id = cs.category_id").
		Where("cs.store_id = ?", storeID).
		Order("cs.sales_value DESC").
		Scan(&data.CategorySales).Error
	if err != nil {
		return fmt.Errorf("query category sales: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadPaymentMix(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("payment_mix pm").
		Select("pm.id, pm.store_id, pm.sales_date, pm.payment_method_id, p.name_th, p.name_en, pm.share").
		Joins("JOIN payment_methods p ON p.id = pm.payment_method_id").
		Where(
			"pm.store_id = ? AND pm.sales_date = (SELECT MAX(pm2.sales_date) FROM payment_mix pm2 WHERE pm2.store_id = ?)",
			storeID,
			storeID,
		).
		Order("pm.share DESC").
		Scan(&data.PaymentMix).Error
	if err != nil {
		return fmt.Errorf("query payment mix: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadTopProducts(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("top_products tp").
		Select("tp.id, tp.store_id, tp.sku, p.name_th, p.name_en, tp.sales_date, tp.sold_quantity, tp.sales_value, tp.trend_percent").
		Joins("JOIN products p ON p.sku = tp.sku").
		Where("tp.store_id = ?", storeID).
		Order("tp.sales_value DESC").
		Scan(&data.TopProducts).Error
	if err != nil {
		return fmt.Errorf("query top products: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadInventoryItems(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("inventory_items ii").
		Select("ii.id, ii.store_id, ii.sku, p.name_th, p.name_en, ii.stock_quantity, ii.reorder_quantity, ii.location_code, ii.price").
		Joins("JOIN products p ON p.sku = ii.sku").
		Where("ii.store_id = ?", storeID).
		Order("ii.location_code, ii.sku").
		Scan(&data.InventoryItems).Error
	if err != nil {
		return fmt.Errorf("query inventory items: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadExpiringInventory(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("expiring_inventory ei").
		Select(`
			ei.id, ei.store_id, ei.sku, p.name_th, p.name_en, ei.category_id, c.name_th AS category_th, c.name_en AS category_en,
			ei.expiry_date, ei.stock_quantity, ei.location_code, ei.price`).
		Joins("JOIN products p ON p.sku = ei.sku").
		Joins("JOIN categories c ON c.id = ei.category_id").
		Where("ei.store_id = ?", storeID).
		Order("ei.expiry_date, ei.stock_quantity DESC").
		Scan(&data.ExpiringInventory).Error
	if err != nil {
		return fmt.Errorf("query expiring inventory: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadLowStockAlerts(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Table("low_stock_alerts lsa").
		Select(`
			lsa.id, lsa.store_id, lsa.sku, p.name_th, p.name_en, lsa.category_id, c.name_th AS category_th, c.name_en AS category_en,
			lsa.stock_quantity, lsa.reorder_quantity, lsa.location_code`).
		Joins("JOIN products p ON p.sku = lsa.sku").
		Joins("JOIN categories c ON c.id = lsa.category_id").
		Where("lsa.store_id = ?", storeID).
		Order("lsa.stock_quantity, lsa.sku").
		Scan(&data.LowStockAlerts).Error
	if err != nil {
		return fmt.Errorf("query low stock alerts: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadDeliveries(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Where("store_id = ?", storeID).
		Order("eta_time, id").
		Find(&data.Deliveries).Error
	if err != nil {
		return fmt.Errorf("query deliveries: %w", err)
	}

	return nil
}

func (repo *gormDashboardRepository) loadSuggestions(ctx context.Context, storeID int64, data *model.DashboardData) error {
	err := repo.db.WithContext(ctx).
		Where("store_id = ?", storeID).
		Order("kind DESC, upside_value DESC").
		Find(&data.Suggestions).Error
	if err != nil {
		return fmt.Errorf("query suggestions: %w", err)
	}

	return nil
}
