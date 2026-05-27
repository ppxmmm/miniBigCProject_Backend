package model

import "time"

// Store contains Mini BigC branch metadata.
type Store struct {
	ID              int64  `json:"id" gorm:"primaryKey"`
	Code            string `json:"code" gorm:"size:32;not null;unique"`
	NameTH          string `json:"name_th" gorm:"size:255;not null"`
	NameEN          string `json:"name_en" gorm:"size:255;not null"`
	ShortNameTH     string `json:"short_name_th" gorm:"size:120;not null"`
	ShortNameEN     string `json:"short_name_en" gorm:"size:120;not null"`
	AddressTH       string `json:"address_th" gorm:"type:text;not null"`
	AddressEN       string `json:"address_en" gorm:"type:text;not null"`
	ManagerNameTH   string `json:"manager_name_th" gorm:"size:160;not null"`
	ManagerNameEN   string `json:"manager_name_en" gorm:"size:160;not null"`
	ManagerInitials string `json:"manager_initials" gorm:"size:12;not null"`
	StaffNameTH     string `json:"staff_name_th" gorm:"size:160;not null"`
	StaffNameEN     string `json:"staff_name_en" gorm:"size:160;not null"`
	StaffInitials   string `json:"staff_initials" gorm:"size:12;not null"`
}

// HourlySale contains current and comparison sales for one hour.
type HourlySale struct {
	ID                   int64     `json:"id"`
	StoreID              int64     `json:"store_id"`
	SaleDate             time.Time `json:"sale_date"`
	Hour                 int       `json:"hour"`
	SalesValue           float64   `json:"sales_value"`
	ComparisonSalesValue float64   `json:"comparison_sales_value"`
}

// DailySale contains current and comparison sales for one day.
type DailySale struct {
	ID                   int64     `json:"id"`
	StoreID              int64     `json:"store_id"`
	SaleDate             time.Time `json:"sale_date"`
	SalesValue           float64   `json:"sales_value"`
	ComparisonSalesValue float64   `json:"comparison_sales_value"`
}

// MonthlySale contains sales for one calendar month.
type MonthlySale struct {
	ID         int64   `json:"id"`
	StoreID    int64   `json:"store_id"`
	Year       int     `json:"year"`
	Month      int     `json:"month"`
	SalesValue float64 `json:"sales_value"`
}

// CategorySale contains category-level revenue summary data.
type CategorySale struct {
	ID           int64     `json:"id"`
	StoreID      int64     `json:"store_id"`
	CategoryID   int64     `json:"category_id"`
	NameTH       string    `json:"name_th"`
	NameEN       string    `json:"name_en"`
	Color        string    `json:"color"`
	SalesDate    time.Time `json:"sales_date"`
	SalesValue   float64   `json:"sales_value"`
	Share        float64   `json:"share"`
	TrendPercent float64   `json:"trend_percent"`
}

// PaymentMix contains one payment method's share for a store.
type PaymentMix struct {
	ID              int64     `json:"id"`
	StoreID         int64     `json:"store_id"`
	SalesDate       time.Time `json:"sales_date"`
	PaymentMethodID int64     `json:"payment_method_id"`
	NameTH          string    `json:"name_th"`
	NameEN          string    `json:"name_en"`
	Share           float64   `json:"share"`
}

// TopProduct contains one top-selling product snapshot.
type TopProduct struct {
	ID           int64     `json:"id"`
	StoreID      int64     `json:"store_id"`
	SKU          string    `json:"sku"`
	NameTH       string    `json:"name_th"`
	NameEN       string    `json:"name_en"`
	SalesDate    time.Time `json:"sales_date"`
	SoldQuantity int       `json:"sold_quantity"`
	SalesValue   float64   `json:"sales_value"`
	TrendPercent float64   `json:"trend_percent"`
}

// InventoryItem contains current product stock state.
type InventoryItem struct {
	ID              int64   `json:"id"`
	StoreID         int64   `json:"store_id"`
	SKU             string  `json:"sku"`
	NameTH          string  `json:"name_th"`
	NameEN          string  `json:"name_en"`
	StockQuantity   int     `json:"stock_quantity"`
	ReorderQuantity int     `json:"reorder_quantity"`
	LocationCode    string  `json:"location_code"`
	Price           float64 `json:"price"`
}

// ExpiringInventoryItem contains product stock approaching expiry.
type ExpiringInventoryItem struct {
	ID            int64     `json:"id"`
	StoreID       int64     `json:"store_id"`
	SKU           string    `json:"sku"`
	NameTH        string    `json:"name_th"`
	NameEN        string    `json:"name_en"`
	CategoryID    int64     `json:"category_id"`
	CategoryTH    string    `json:"category_th"`
	CategoryEN    string    `json:"category_en"`
	ExpiryDate    time.Time `json:"expiry_date"`
	StockQuantity int       `json:"stock_quantity"`
	LocationCode  string    `json:"location_code"`
	Price         float64   `json:"price"`
}

// LowStockAlert contains product stock below reorder threshold.
type LowStockAlert struct {
	ID              int64  `json:"id"`
	StoreID         int64  `json:"store_id"`
	SKU             string `json:"sku"`
	NameTH          string `json:"name_th"`
	NameEN          string `json:"name_en"`
	CategoryID      int64  `json:"category_id"`
	CategoryTH      string `json:"category_th"`
	CategoryEN      string `json:"category_en"`
	StockQuantity   int    `json:"stock_quantity"`
	ReorderQuantity int    `json:"reorder_quantity"`
	LocationCode    string `json:"location_code"`
}

// Delivery contains delivery order and fulfillment status.
type Delivery struct {
	ID             string  `json:"id" gorm:"primaryKey;size:40"`
	StoreID        int64   `json:"store_id" gorm:"not null;index:idx_deliveries_store_status;index:idx_deliveries_store_eta"`
	CustomerNameTH string  `json:"customer_name_th" gorm:"size:160;not null"`
	CustomerNameEN string  `json:"customer_name_en" gorm:"size:160;not null"`
	AddressTH      string  `json:"address_th" gorm:"type:text;not null"`
	AddressEN      string  `json:"address_en" gorm:"type:text;not null"`
	ItemCount      int     `json:"item_count" gorm:"not null"`
	OrderValue     float64 `json:"order_value" gorm:"type:numeric(12,2);not null"`
	DriverNameTH   string  `json:"driver_name_th" gorm:"size:160;not null"`
	DriverNameEN   string  `json:"driver_name_en" gorm:"size:160;not null"`
	Status         string  `json:"status" gorm:"size:40;not null;index:idx_deliveries_store_status"`
	ETATime        string  `json:"eta_time" gorm:"type:time;not null;index:idx_deliveries_store_eta"`
	IsLate         bool    `json:"is_late" gorm:"not null"`
	DistanceKM     float64 `json:"distance_km" gorm:"type:numeric(8,2);not null"`
}

// Suggestion contains dashboard recommendations such as promotions, events,
// operations, inventory alerts, risks, and customer insights.
type Suggestion struct {
	ID            string  `json:"id" gorm:"primaryKey;size:40"`
	StoreID       int64   `json:"store_id" gorm:"not null;index:idx_suggestions_store_kind_type"`
	Kind          string  `json:"kind" gorm:"size:40;not null;check:kind IN ('promo','event','operation','inventory','risk','customer');index:idx_suggestions_store_kind_type"`
	Icon          string  `json:"icon" gorm:"size:40;not null"`
	TitleTH       string  `json:"title_th" gorm:"size:255;not null"`
	TitleEN       string  `json:"title_en" gorm:"size:255;not null"`
	DescriptionTH string  `json:"description_th" gorm:"type:text;not null"`
	DescriptionEN string  `json:"description_en" gorm:"type:text;not null"`
	UpsideValue   float64 `json:"upside_value" gorm:"type:numeric(14,2);not null"`
	Confidence    float64 `json:"confidence" gorm:"type:numeric(7,4);not null"`
	DurationTH    string  `json:"duration_th" gorm:"size:120;not null"`
	DurationEN    string  `json:"duration_en" gorm:"size:120;not null"`
	TargetTH      string  `json:"target_th" gorm:"size:160;not null"`
	TargetEN      string  `json:"target_en" gorm:"size:160;not null"`
	Type          string  `json:"type" gorm:"size:40;not null;index:idx_suggestions_store_kind_type"`
}

// DashboardData is the aggregate response consumed by the frontend dashboard.
type DashboardData struct {
	Store             Store                   `json:"store"`
	HourlySales       []HourlySale            `json:"hourly_sales"`
	DailySales        []DailySale             `json:"daily_sales"`
	MonthlySales      []MonthlySale           `json:"monthly_sales"`
	CategorySales     []CategorySale          `json:"category_sales"`
	PaymentMix        []PaymentMix            `json:"payment_mix"`
	TopProducts       []TopProduct            `json:"top_products"`
	InventoryItems    []InventoryItem         `json:"inventory_items"`
	ExpiringInventory []ExpiringInventoryItem `json:"expiring_inventory"`
	LowStockAlerts    []LowStockAlert         `json:"low_stock_alerts"`
	Deliveries        []Delivery              `json:"deliveries"`
	Suggestions       []Suggestion            `json:"suggestions"`
}
