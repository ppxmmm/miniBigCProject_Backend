package model

import "time"

// Category stores product or reporting category metadata.
type Category struct {
	ID     int64  `json:"id" gorm:"primaryKey"`
	NameTH string `json:"name_th" gorm:"size:160;not null"`
	NameEN string `json:"name_en" gorm:"size:160;not null;unique"`
	Color  string `json:"color" gorm:"size:80;not null;default:''"`
}

// PaymentMethod stores supported payment method labels.
type PaymentMethod struct {
	ID     int64  `json:"id" gorm:"primaryKey"`
	NameTH string `json:"name_th" gorm:"size:160;not null"`
	NameEN string `json:"name_en" gorm:"size:160;not null;unique"`
}

// Product stores product master data.
type ProductEntity struct {
	SKU        string `json:"sku" gorm:"primaryKey;size:40"`
	NameTH     string `json:"name_th" gorm:"size:255;not null"`
	NameEN     string `json:"name_en" gorm:"size:255;not null"`
	CategoryID int64  `json:"category_id"`
}

// TableName maps ProductEntity to the products table.
func (ProductEntity) TableName() string {
	return "products"
}

// SalesHourly stores current and comparison hourly sales.
type SalesHourly struct {
	ID                   int64     `json:"id" gorm:"primaryKey"`
	StoreID              int64     `json:"store_id" gorm:"not null"`
	SaleDate             time.Time `json:"sale_date" gorm:"type:date;not null"`
	Hour                 int       `json:"hour" gorm:"not null;check:hour >= 0 AND hour <= 23"`
	SalesValue           float64   `json:"sales_value" gorm:"type:numeric(14,2);not null"`
	ComparisonSalesValue float64   `json:"comparison_sales_value" gorm:"type:numeric(14,2);not null"`
}

// TableName maps SalesHourly to the sales_hourly table.
func (SalesHourly) TableName() string {
	return "sales_hourly"
}

// SalesDaily stores current and comparison daily sales.
type SalesDaily struct {
	ID                   int64     `json:"id" gorm:"primaryKey"`
	StoreID              int64     `json:"store_id" gorm:"not null"`
	SaleDate             time.Time `json:"sale_date" gorm:"type:date;not null"`
	SalesValue           float64   `json:"sales_value" gorm:"type:numeric(14,2);not null"`
	ComparisonSalesValue float64   `json:"comparison_sales_value" gorm:"type:numeric(14,2);not null"`
}

// TableName maps SalesDaily to the sales_daily table.
func (SalesDaily) TableName() string {
	return "sales_daily"
}

// SalesMonthly stores monthly sales values.
type SalesMonthly struct {
	ID         int64   `json:"id" gorm:"primaryKey"`
	StoreID    int64   `json:"store_id" gorm:"not null"`
	Year       int     `json:"year" gorm:"not null"`
	Month      int     `json:"month" gorm:"not null;check:month >= 1 AND month <= 12"`
	SalesValue float64 `json:"sales_value" gorm:"type:numeric(14,2);not null"`
}

// TableName maps SalesMonthly to the sales_monthly table.
func (SalesMonthly) TableName() string {
	return "sales_monthly"
}

// CategorySalesEntity stores category-level sales summary data.
type CategorySalesEntity struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	StoreID      int64     `json:"store_id" gorm:"not null"`
	CategoryID   int64     `json:"category_id" gorm:"not null"`
	SalesDate    time.Time `json:"sales_date" gorm:"type:date;not null"`
	SalesValue   float64   `json:"sales_value" gorm:"type:numeric(14,2);not null"`
	Share        float64   `json:"share" gorm:"type:numeric(7,4);not null"`
	TrendPercent float64   `json:"trend_percent" gorm:"type:numeric(7,2);not null"`
}

// TableName maps CategorySalesEntity to the category_sales table.
func (CategorySalesEntity) TableName() string {
	return "category_sales"
}

// PaymentMixEntity stores payment method share data.
type PaymentMixEntity struct {
	ID              int64     `json:"id" gorm:"primaryKey"`
	StoreID         int64     `json:"store_id" gorm:"not null"`
	SalesDate       time.Time `json:"sales_date" gorm:"type:date;not null"`
	PaymentMethodID int64     `json:"payment_method_id" gorm:"not null"`
	Share           float64   `json:"share" gorm:"type:numeric(7,4);not null"`
}

// TableName maps PaymentMixEntity to the payment_mix table.
func (PaymentMixEntity) TableName() string {
	return "payment_mix"
}

// TopProductEntity stores top-selling product snapshots.
type TopProductEntity struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	StoreID      int64     `json:"store_id" gorm:"not null;index:idx_top_products_store_date"`
	SKU          string    `json:"sku" gorm:"size:40;not null"`
	SalesDate    time.Time `json:"sales_date" gorm:"type:date;not null;index:idx_top_products_store_date"`
	SoldQuantity int       `json:"sold_quantity" gorm:"not null"`
	SalesValue   float64   `json:"sales_value" gorm:"type:numeric(14,2);not null"`
	TrendPercent float64   `json:"trend_percent" gorm:"type:numeric(7,2);not null"`
}

// TableName maps TopProductEntity to the top_products table.
func (TopProductEntity) TableName() string {
	return "top_products"
}

// InventoryItemEntity stores current inventory state.
type InventoryItemEntity struct {
	ID              int64   `json:"id" gorm:"primaryKey"`
	StoreID         int64   `json:"store_id" gorm:"not null"`
	SKU             string  `json:"sku" gorm:"size:40;not null"`
	StockQuantity   int     `json:"stock_quantity" gorm:"not null"`
	ReorderQuantity int     `json:"reorder_quantity" gorm:"not null"`
	LocationCode    string  `json:"location_code" gorm:"size:40;not null"`
	Price           float64 `json:"price" gorm:"type:numeric(12,2);not null"`
}

// TableName maps InventoryItemEntity to the inventory_items table.
func (InventoryItemEntity) TableName() string {
	return "inventory_items"
}

// ExpiringInventoryEntity stores products approaching expiry.
type ExpiringInventoryEntity struct {
	ID            int64     `json:"id" gorm:"primaryKey"`
	StoreID       int64     `json:"store_id" gorm:"not null;index:idx_expiring_inventory_store_expiry"`
	SKU           string    `json:"sku" gorm:"size:40;not null"`
	CategoryID    int64     `json:"category_id" gorm:"not null"`
	ExpiryDate    time.Time `json:"expiry_date" gorm:"type:date;not null;index:idx_expiring_inventory_store_expiry"`
	StockQuantity int       `json:"stock_quantity" gorm:"not null"`
	LocationCode  string    `json:"location_code" gorm:"size:40;not null"`
	Price         float64   `json:"price" gorm:"type:numeric(12,2);not null"`
}

// TableName maps ExpiringInventoryEntity to the expiring_inventory table.
func (ExpiringInventoryEntity) TableName() string {
	return "expiring_inventory"
}

// LowStockAlertEntity stores products below reorder threshold.
type LowStockAlertEntity struct {
	ID              int64  `json:"id" gorm:"primaryKey"`
	StoreID         int64  `json:"store_id" gorm:"not null;index:idx_low_stock_alerts_store_stock"`
	SKU             string `json:"sku" gorm:"size:40;not null"`
	CategoryID      int64  `json:"category_id" gorm:"not null"`
	StockQuantity   int    `json:"stock_quantity" gorm:"not null;index:idx_low_stock_alerts_store_stock"`
	ReorderQuantity int    `json:"reorder_quantity" gorm:"not null"`
	LocationCode    string `json:"location_code" gorm:"size:40;not null"`
}

// TableName maps LowStockAlertEntity to the low_stock_alerts table.
func (LowStockAlertEntity) TableName() string {
	return "low_stock_alerts"
}
