package db

import (
	"context"
	"fmt"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/util"
	"gorm.io/gorm"
)

// Bootstrap migrates the dashboard schema and loads mock data when missing.
func Bootstrap(ctx context.Context, database *gorm.DB) error {
	if err := migrateMissingTables(ctx, database); err != nil {
		return fmt.Errorf("auto migrate dashboard schema: %w", err)
	}
	if err := ensureIndexes(ctx, database); err != nil {
		return fmt.Errorf("ensure dashboard indexes: %w", err)
	}

	if err := util.SeedMockData(ctx, database); err != nil {
		return fmt.Errorf("seed mock data: %w", err)
	}

	return nil
}

func migrateMissingTables(ctx context.Context, database *gorm.DB) error {
	models := []any{
		&model.Store{},
		&model.Category{},
		&model.PaymentMethod{},
		&model.ProductEntity{},
		&model.SalesHourly{},
		&model.SalesDaily{},
		&model.SalesMonthly{},
		&model.CategorySalesEntity{},
		&model.PaymentMixEntity{},
		&model.TopProductEntity{},
		&model.InventoryItemEntity{},
		&model.ExpiringInventoryEntity{},
		&model.LowStockAlertEntity{},
		&model.Delivery{},
		&model.Suggestion{},
	}

	tx := database.WithContext(ctx)
	for _, migrationModel := range models {
		if err := tx.AutoMigrate(migrationModel); err != nil {
			return fmt.Errorf("auto migrate table: %w", err)
		}
	}

	return nil
}

func ensureIndexes(ctx context.Context, database *gorm.DB) error {
	statements := []string{
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_stores_code ON stores(code)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_name_en ON categories(name_en)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_methods_name_en ON payment_methods(name_en)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_hourly_store_date_hour ON sales_hourly(store_id, sale_date, hour)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_daily_store_date ON sales_daily(store_id, sale_date)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_sales_monthly_store_year_month ON sales_monthly(store_id, year, month)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_category_sales_store_category_date ON category_sales(store_id, category_id, sales_date)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_payment_mix_store_method_date ON payment_mix(store_id, payment_method_id, sales_date)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_top_products_store_sku_date ON top_products(store_id, sku, sales_date)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_inventory_store_sku_location ON inventory_items(store_id, sku, location_code)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_expiring_store_sku_expiry_location ON expiring_inventory(store_id, sku, expiry_date, location_code)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_low_stock_store_sku_location ON low_stock_alerts(store_id, sku, location_code)`,
	}

	for _, statement := range statements {
		if err := database.WithContext(ctx).Exec(statement).Error; err != nil {
			return fmt.Errorf("execute index statement: %w", err)
		}
	}

	return nil
}
