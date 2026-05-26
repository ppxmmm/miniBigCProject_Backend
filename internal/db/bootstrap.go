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
	if err := migrateColumnTypes(ctx, database); err != nil {
		return fmt.Errorf("migrate column types: %w", err)
	}
	if err := ensureConstraints(ctx, database); err != nil {
		return fmt.Errorf("ensure dashboard constraints: %w", err)
	}
	if err := ensureIndexes(ctx, database); err != nil {
		return fmt.Errorf("ensure dashboard indexes: %w", err)
	}

	if err := util.SeedMockData(ctx, database); err != nil {
		return fmt.Errorf("seed mock data: %w", err)
	}

	return nil
}

// migrateColumnTypes fixes column types that AutoMigrate cannot change on existing tables.
func migrateColumnTypes(ctx context.Context, database *gorm.DB) error {
	// eta_time may have been created as timestamptz when ETATime was time.Time;
	// the model now declares type:time, so coerce it here idempotently.
	stmt := `
		DO $$ BEGIN
			IF (SELECT data_type FROM information_schema.columns
				WHERE table_name = 'deliveries' AND column_name = 'eta_time')
				= 'timestamp with time zone' THEN
				ALTER TABLE deliveries ALTER COLUMN eta_time TYPE time
					USING eta_time::time;
			END IF;
		END $$`
	if err := database.WithContext(ctx).Exec(stmt).Error; err != nil {
		return fmt.Errorf("alter deliveries.eta_time to time: %w", err)
	}
	return nil
}

func ensureConstraints(ctx context.Context, database *gorm.DB) error {
	statements := []string{
		`
		ALTER TABLE deliveries
		DROP CONSTRAINT IF EXISTS chk_deliveries_status;
		`,
		`
		ALTER TABLE deliveries
		ADD CONSTRAINT chk_deliveries_status
		CHECK (status IN ('preparing', 'enRoute', 'delivered'));
		`,

		// Drop old suggestion kind constraints.
		// GORM/Postgres may create this name automatically: suggestions_kind_check
		// Your custom migration may create this name: chk_suggestions_kind
		`
		ALTER TABLE suggestions
		DROP CONSTRAINT IF EXISTS chk_suggestions_kind;
		`,
		`
		ALTER TABLE suggestions
		DROP CONSTRAINT IF EXISTS suggestions_kind_check;
		`,

		// Add the new correct constraint.
		`
		ALTER TABLE suggestions
		ADD CONSTRAINT chk_suggestions_kind
		CHECK (kind IN ('promo', 'event', 'operation', 'inventory', 'risk', 'customer'));
		`,

		// Optional: fix suggestion type constraint too.
		`
		ALTER TABLE suggestions
		DROP CONSTRAINT IF EXISTS chk_suggestions_type;
		`,
		`
		ALTER TABLE suggestions
		DROP CONSTRAINT IF EXISTS suggestions_type_check;
		`,
		`
		ALTER TABLE suggestions
		ADD CONSTRAINT chk_suggestions_type
		CHECK (
			type IN (
				'markdown',
				'bundle',
				'discount',
				'event',
				'operation',
				'staffing',
				'delivery',
				'reorder',
				'quality',
				'fifo',
				'loss_prevention',
				'local_promo',
				'assortment',
				'merchandising'
			)
		);
		`,
	}

	for _, statement := range statements {
		if err := database.WithContext(ctx).Exec(statement).Error; err != nil {
			return fmt.Errorf("execute constraint statement: %w", err)
		}
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
		`CREATE INDEX IF NOT EXISTS idx_deliveries_store_status ON deliveries(store_id, status)`,
		`CREATE INDEX IF NOT EXISTS idx_deliveries_store_eta ON deliveries(store_id, eta_time)`,
		`CREATE INDEX IF NOT EXISTS idx_suggestions_store_kind_type ON suggestions(store_id, kind, type)`,
	}

	for _, statement := range statements {
		if err := database.WithContext(ctx).Exec(statement).Error; err != nil {
			return fmt.Errorf("execute index statement: %w", err)
		}
	}

	return nil
}
