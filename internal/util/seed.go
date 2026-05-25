package util

import (
	"context"
	"fmt"
	"time"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SeedMockData loads dashboard mock data using idempotent upserts.
func SeedMockData(ctx context.Context, database *gorm.DB) error {
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		store := model.Store{
			Code:            "MBC-0421",
			NameTH:          "มินิ บิ๊กซี สาขาทองหล่อ ซอย 13",
			NameEN:          "Mini BigC · Thonglor Soi 13",
			ShortNameTH:     "ทองหล่อ ซ.13",
			ShortNameEN:     "Thonglor 13",
			AddressTH:       "121/4 ซ.สุขุมวิท 55, แขวงคลองตันเหนือ, วัฒนา, กรุงเทพฯ 10110",
			AddressEN:       "121/4 Sukhumvit 55, Khlong Tan Nuea, Watthana, Bangkok 10110",
			ManagerNameTH:   "ปริญญา ทวีศักดิ์",
			ManagerNameEN:   "Parinya Taweesak",
			ManagerInitials: "PT",
			StaffNameTH:     "ณัฐวุฒิ สมบูรณ์",
			StaffNameEN:     "Nattawut Somboon",
			StaffInitials:   "NS",
		}
		if err := tx.Clauses(upsertByColumns([]string{"code"})).Create(&store).Error; err != nil {
			return fmt.Errorf("upsert store: %w", err)
		}
		if err := tx.Where("code = ?", store.Code).First(&store).Error; err != nil {
			return fmt.Errorf("query seeded store: %w", err)
		}

		categoryIDs, err := seedCategories(tx)
		if err != nil {
			return err
		}
		paymentMethodIDs, err := seedPaymentMethods(tx)
		if err != nil {
			return err
		}
		if err := seedProducts(tx, categoryIDs); err != nil {
			return err
		}

		today := time.Now().In(time.FixedZone("Asia/Bangkok", 7*60*60)).Truncate(24 * time.Hour)
		seeders := []func(*gorm.DB, int64, time.Time, map[string]int64, map[string]int64) error{
			seedSales,
			seedCategorySales,
			seedPaymentMix,
			seedTopProducts,
			seedInventory,
			seedDeliveries,
			seedSuggestions,
		}
		for _, seeder := range seeders {
			if err := seeder(tx, store.ID, today, categoryIDs, paymentMethodIDs); err != nil {
				return err
			}
		}

		return nil
	})
}

func seedCategories(tx *gorm.DB) (map[string]int64, error) {
	categories := []model.Category{
		{NameTH: "อาหารและเครื่องดื่ม", NameEN: "Food & Beverage", Color: "oklch(0.58 0.16 150)"},
		{NameTH: "ของใช้ในบ้าน", NameEN: "Household", Color: "oklch(0.62 0.14 200)"},
		{NameTH: "ของใช้ส่วนตัว", NameEN: "Personal Care", Color: "oklch(0.62 0.14 280)"},
		{NameTH: "ขนมและของว่าง", NameEN: "Snacks", Color: "oklch(0.72 0.15 65)"},
		{NameTH: "อาหารแช่แข็ง", NameEN: "Frozen", Color: "oklch(0.66 0.14 320)"},
		{NameTH: "อื่น ๆ", NameEN: "Other", Color: "oklch(0.70 0.04 95)"},
		{NameTH: "อาหาร", NameEN: "Food", Color: "oklch(0.58 0.16 150)"},
		{NameTH: "แช่เย็น", NameEN: "Chilled", Color: "oklch(0.62 0.14 200)"},
		{NameTH: "ผัก-ผลไม้", NameEN: "Produce", Color: "oklch(0.58 0.16 130)"},
		{NameTH: "ของใช้ส่วนตัว", NameEN: "Personal", Color: "oklch(0.62 0.14 280)"},
		{NameTH: "ของใช้บ้าน", NameEN: "Household Items", Color: "oklch(0.62 0.14 200)"},
		{NameTH: "เครื่องดื่ม", NameEN: "Beverage", Color: "oklch(0.58 0.16 150)"},
	}
	if err := tx.Clauses(upsertByColumns([]string{"name_en"})).Create(&categories).Error; err != nil {
		return nil, fmt.Errorf("upsert categories: %w", err)
	}

	ids := make(map[string]int64, len(categories))
	for _, category := range categories {
		var found model.Category
		if err := tx.Where("name_en = ?", category.NameEN).First(&found).Error; err != nil {
			return nil, fmt.Errorf("query category %s: %w", category.NameEN, err)
		}
		ids[found.NameEN] = found.ID
	}

	return ids, nil
}

func seedPaymentMethods(tx *gorm.DB) (map[string]int64, error) {
	methods := []model.PaymentMethod{
		{NameTH: "เงินสด", NameEN: "Cash"},
		{NameTH: "พร้อมเพย์ / QR", NameEN: "PromptPay / QR"},
		{NameTH: "บัตรเครดิต", NameEN: "Credit card"},
		{NameTH: "เดบิต", NameEN: "Debit"},
		{NameTH: "True Money / Wallet", NameEN: "True Money / Wallet"},
	}
	if err := tx.Clauses(upsertByColumns([]string{"name_en"})).Create(&methods).Error; err != nil {
		return nil, fmt.Errorf("upsert payment methods: %w", err)
	}

	ids := make(map[string]int64, len(methods))
	for _, method := range methods {
		var found model.PaymentMethod
		if err := tx.Where("name_en = ?", method.NameEN).First(&found).Error; err != nil {
			return nil, fmt.Errorf("query payment method %s: %w", method.NameEN, err)
		}
		ids[found.NameEN] = found.ID
	}

	return ids, nil
}

func seedProducts(tx *gorm.DB, categoryIDs map[string]int64) error {
	products := []model.ProductEntity{
		{SKU: "FB-0102", NameTH: "ลีโอ เบียร์ กระป๋อง 320 มล.", NameEN: "Leo beer can 320ml", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FB-2284", NameTH: "เลย์ คลาสสิค 75 ก.", NameEN: "Lays Classic 75g", CategoryID: categoryIDs["Snacks"]},
		{SKU: "FB-0411", NameTH: "นมไทย-เดนมาร์ค จืด 200 มล.", NameEN: "Foremost milk plain 200ml", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-0099", NameTH: "น้ำดื่ม คริสตัล 600 มล. (6 ขวด)", NameEN: "Crystal water 600ml × 6", CategoryID: categoryIDs["Beverage"]},
		{SKU: "PC-0331", NameTH: "หน้ากากอนามัย 4 ชั้น (50 ชิ้น)", NameEN: "Surgical mask 4-ply × 50", CategoryID: categoryIDs["Personal"]},
		{SKU: "HH-1820", NameTH: "ผงซักฟอก เปา 800 ก.", NameEN: "Pao detergent 800g", CategoryID: categoryIDs["Household Items"]},
		{SKU: "FB-1140", NameTH: "ขนมปังฟาร์มเฮ้าส์ โฮลวีท", NameEN: "Farmhouse wholewheat bread", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-0822", NameTH: "ไส้กรอกบีทาโกร 250 ก.", NameEN: "Betagro sausage 250g", CategoryID: categoryIDs["Chilled"]},
		{SKU: "FB-1990", NameTH: "โยเกิร์ตดัชชี่ รสกล้วยหอม", NameEN: "Dutchie banana yoghurt", CategoryID: categoryIDs["Chilled"]},
		{SKU: "FB-2210", NameTH: "ส้มสายน้ำผึ้ง 1 กก.", NameEN: "Honey oranges 1kg", CategoryID: categoryIDs["Produce"]},
		{SKU: "FB-3084", NameTH: "เต้าหู้ขาวอ่อน คาวบอย", NameEN: "Cowboy soft tofu", CategoryID: categoryIDs["Chilled"]},
		{SKU: "FB-0188", NameTH: "น้ำพริกแม่ประนอม", NameEN: "Mae Pranom chilli paste", CategoryID: categoryIDs["Food"]},
		{SKU: "PC-0470", NameTH: "ผ้าอนามัยลอริเอะ ขนาดกลาง", NameEN: "Laurier sanitary pads M", CategoryID: categoryIDs["Personal"]},
		{SKU: "HH-2201", NameTH: "กระดาษทิชชู่ Cellox 12 ม้วน", NameEN: "Cellox tissue × 12", CategoryID: categoryIDs["Household Items"]},
	}
	if err := tx.Clauses(upsertByColumns([]string{"sku"})).Create(&products).Error; err != nil {
		return fmt.Errorf("upsert products: %w", err)
	}

	return nil
}

func seedSales(tx *gorm.DB, storeID int64, today time.Time, _ map[string]int64, _ map[string]int64) error {
	hourlyValues := []float64{3200, 4100, 5800, 8400, 11200, 9800, 12400, 15600, 14200, 11800, 9400, 13200, 18900, 21400, 19200, 14800, 11600, 8200}
	hourlyComparison := []float64{2900, 3800, 5200, 7900, 10400, 9100, 11600, 14200, 13100, 10800, 8900, 12100, 17400, 19800, 17600, 13900, 10900, 7800}
	hourly := make([]model.SalesHourly, 0, len(hourlyValues))
	for index, value := range hourlyValues {
		hourly = append(hourly, model.SalesHourly{
			StoreID:              storeID,
			SaleDate:             today,
			Hour:                 index + 6,
			SalesValue:           value,
			ComparisonSalesValue: hourlyComparison[index],
		})
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sale_date", "hour"})).Create(&hourly).Error; err != nil {
		return fmt.Errorf("upsert hourly sales: %w", err)
	}

	dailyValues := []float64{184000, 168000, 192000, 211000, 248000, 296000, 264000}
	dailyComparison := []float64{172000, 161000, 178000, 199000, 226000, 274000, 248000}
	daily := make([]model.SalesDaily, 0, len(dailyValues))
	for index, value := range dailyValues {
		daily = append(daily, model.SalesDaily{
			StoreID:              storeID,
			SaleDate:             today.AddDate(0, 0, index-len(dailyValues)+1),
			SalesValue:           value,
			ComparisonSalesValue: dailyComparison[index],
		})
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sale_date"})).Create(&daily).Error; err != nil {
		return fmt.Errorf("upsert daily sales: %w", err)
	}

	monthlyValues := []float64{5420000, 5180000, 5960000, 6180000, 6740000, 6920000, 7180000, 7040000, 6880000, 7320000, 7860000, 8140000}
	monthly := make([]model.SalesMonthly, 0, len(monthlyValues))
	for index, value := range monthlyValues {
		monthly = append(monthly, model.SalesMonthly{
			StoreID:    storeID,
			Year:       today.Year(),
			Month:      index + 1,
			SalesValue: value,
		})
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "year", "month"})).Create(&monthly).Error; err != nil {
		return fmt.Errorf("upsert monthly sales: %w", err)
	}

	return nil
}

func seedCategorySales(tx *gorm.DB, storeID int64, today time.Time, categoryIDs map[string]int64, _ map[string]int64) error {
	rows := []model.CategorySalesEntity{
		{StoreID: storeID, CategoryID: categoryIDs["Food & Beverage"], SalesDate: today, SalesValue: 142800, Share: 0.38, TrendPercent: 6.2},
		{StoreID: storeID, CategoryID: categoryIDs["Household"], SalesDate: today, SalesValue: 78400, Share: 0.21, TrendPercent: -2.1},
		{StoreID: storeID, CategoryID: categoryIDs["Personal Care"], SalesDate: today, SalesValue: 56200, Share: 0.15, TrendPercent: 4.8},
		{StoreID: storeID, CategoryID: categoryIDs["Snacks"], SalesDate: today, SalesValue: 49100, Share: 0.13, TrendPercent: 11.3},
		{StoreID: storeID, CategoryID: categoryIDs["Frozen"], SalesDate: today, SalesValue: 28400, Share: 0.075, TrendPercent: -1.4},
		{StoreID: storeID, CategoryID: categoryIDs["Other"], SalesDate: today, SalesValue: 21100, Share: 0.055, TrendPercent: 2.0},
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "category_id", "sales_date"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert category sales: %w", err)
	}

	return nil
}

func seedPaymentMix(tx *gorm.DB, storeID int64, today time.Time, _ map[string]int64, paymentMethodIDs map[string]int64) error {
	rows := []model.PaymentMixEntity{
		{StoreID: storeID, SalesDate: today, PaymentMethodID: paymentMethodIDs["Cash"], Share: 0.38},
		{StoreID: storeID, SalesDate: today, PaymentMethodID: paymentMethodIDs["PromptPay / QR"], Share: 0.34},
		{StoreID: storeID, SalesDate: today, PaymentMethodID: paymentMethodIDs["Credit card"], Share: 0.18},
		{StoreID: storeID, SalesDate: today, PaymentMethodID: paymentMethodIDs["Debit"], Share: 0.07},
		{StoreID: storeID, SalesDate: today, PaymentMethodID: paymentMethodIDs["True Money / Wallet"], Share: 0.03},
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "payment_method_id", "sales_date"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert payment mix: %w", err)
	}

	return nil
}

func seedTopProducts(tx *gorm.DB, storeID int64, today time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.TopProductEntity{
		{StoreID: storeID, SKU: "FB-0102", SalesDate: today, SoldQuantity: 184, SalesValue: 9568, TrendPercent: 12},
		{StoreID: storeID, SKU: "FB-2284", SalesDate: today, SoldQuantity: 162, SalesValue: 4374, TrendPercent: 8},
		{StoreID: storeID, SKU: "FB-0411", SalesDate: today, SoldQuantity: 148, SalesValue: 2516, TrendPercent: 5},
		{StoreID: storeID, SKU: "FB-0099", SalesDate: today, SoldQuantity: 121, SalesValue: 8470, TrendPercent: -2},
		{StoreID: storeID, SKU: "PC-0331", SalesDate: today, SoldQuantity: 96, SalesValue: 11520, TrendPercent: 18},
		{StoreID: storeID, SKU: "HH-1820", SalesDate: today, SoldQuantity: 88, SalesValue: 7392, TrendPercent: 3},
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sku", "sales_date"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert top products: %w", err)
	}

	return nil
}

func seedInventory(tx *gorm.DB, storeID int64, today time.Time, categoryIDs map[string]int64, _ map[string]int64) error {
	expiring := []struct {
		entity model.ExpiringInventoryEntity
		price  float64
	}{
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-0411", CategoryID: categoryIDs["Food"], ExpiryDate: today.AddDate(0, 0, 1), StockQuantity: 24, LocationCode: "F-12-B", Price: 17}, 17},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-1140", CategoryID: categoryIDs["Food"], ExpiryDate: today.AddDate(0, 0, 2), StockQuantity: 11, LocationCode: "A-04-C", Price: 45}, 45},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-0822", CategoryID: categoryIDs["Chilled"], ExpiryDate: today.AddDate(0, 0, 3), StockQuantity: 18, LocationCode: "C-01-A", Price: 89}, 89},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-1990", CategoryID: categoryIDs["Chilled"], ExpiryDate: today, StockQuantity: 6, LocationCode: "C-02-A", Price: 19}, 19},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-2210", CategoryID: categoryIDs["Produce"], ExpiryDate: today.AddDate(0, 0, 2), StockQuantity: 14, LocationCode: "P-01-A", Price: 79}, 79},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-3084", CategoryID: categoryIDs["Chilled"], ExpiryDate: today.AddDate(0, 0, 4), StockQuantity: 22, LocationCode: "C-03-B", Price: 18}, 18},
		{model.ExpiringInventoryEntity{StoreID: storeID, SKU: "FB-0188", CategoryID: categoryIDs["Food"], ExpiryDate: today.AddDate(0, 0, 6), StockQuantity: 9, LocationCode: "A-09-D", Price: 32}, 32},
	}
	expiringEntities := make([]model.ExpiringInventoryEntity, 0, len(expiring))
	inventory := make([]model.InventoryItemEntity, 0, 13)
	for _, row := range expiring {
		expiringEntities = append(expiringEntities, row.entity)
		inventory = append(inventory, model.InventoryItemEntity{
			StoreID:         storeID,
			SKU:             row.entity.SKU,
			StockQuantity:   row.entity.StockQuantity,
			ReorderQuantity: 10,
			LocationCode:    row.entity.LocationCode,
			Price:           row.price,
		})
	}

	lowStock := []model.LowStockAlertEntity{
		{StoreID: storeID, SKU: "PC-0331", CategoryID: categoryIDs["Personal"], StockQuantity: 4, ReorderQuantity: 30, LocationCode: "G-02-A"},
		{StoreID: storeID, SKU: "HH-1820", CategoryID: categoryIDs["Household Items"], StockQuantity: 7, ReorderQuantity: 24, LocationCode: "H-04-B"},
		{StoreID: storeID, SKU: "FB-0099", CategoryID: categoryIDs["Beverage"], StockQuantity: 12, ReorderQuantity: 36, LocationCode: "B-01-A"},
		{StoreID: storeID, SKU: "PC-0470", CategoryID: categoryIDs["Personal"], StockQuantity: 5, ReorderQuantity: 18, LocationCode: "G-05-B"},
		{StoreID: storeID, SKU: "FB-2284", CategoryID: categoryIDs["Snacks"], StockQuantity: 18, ReorderQuantity: 40, LocationCode: "S-02-A"},
		{StoreID: storeID, SKU: "HH-2201", CategoryID: categoryIDs["Household Items"], StockQuantity: 3, ReorderQuantity: 15, LocationCode: "H-06-C"},
	}
	lowStockPrices := map[string]float64{"PC-0331": 120, "HH-1820": 84, "FB-0099": 70, "PC-0470": 59, "FB-2284": 27, "HH-2201": 159}
	for _, row := range lowStock {
		inventory = append(inventory, model.InventoryItemEntity{
			StoreID:         storeID,
			SKU:             row.SKU,
			StockQuantity:   row.StockQuantity,
			ReorderQuantity: row.ReorderQuantity,
			LocationCode:    row.LocationCode,
			Price:           lowStockPrices[row.SKU],
		})
	}

	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sku", "location_code"})).Create(&inventory).Error; err != nil {
		return fmt.Errorf("upsert inventory items: %w", err)
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sku", "expiry_date", "location_code"})).Create(&expiringEntities).Error; err != nil {
		return fmt.Errorf("upsert expiring inventory: %w", err)
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sku", "location_code"})).Create(&lowStock).Error; err != nil {
		return fmt.Errorf("upsert low stock alerts: %w", err)
	}

	return nil
}

func seedDeliveries(tx *gorm.DB, storeID int64, _ time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.Delivery{
		{ID: "BC-26052201", StoreID: storeID, CustomerNameTH: "คุณณัฐกานต์ ม.", CustomerNameEN: "K. Natthakarn M.", AddressTH: "ทองหล่อ ซ.10", AddressEN: "Thonglor Soi 10", ItemCount: 8, OrderValue: 487, DriverNameTH: "วินัย", DriverNameEN: "Winai", Status: "enRoute", ETATime: "14:25", IsLate: false, DistanceKM: 1.2},
		{ID: "BC-26052202", StoreID: storeID, CustomerNameTH: "คุณปานทิพย์ ส.", CustomerNameEN: "K. Panthip S.", AddressTH: "เอกมัย ซ.12", AddressEN: "Ekkamai Soi 12", ItemCount: 22, OrderValue: 1284, DriverNameTH: "สมชาย", DriverNameEN: "Somchai", Status: "enRoute", ETATime: "14:35", IsLate: true, DistanceKM: 2.4},
		{ID: "BC-26052203", StoreID: storeID, CustomerNameTH: "คุณภาสกร ว.", CustomerNameEN: "K. Phasakorn W.", AddressTH: "ทองหล่อ ซ.5", AddressEN: "Thonglor Soi 5", ItemCount: 14, OrderValue: 892, DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong", Status: "preparing", ETATime: "15:00", IsLate: false, DistanceKM: 0.8},
		{ID: "BC-26052204", StoreID: storeID, CustomerNameTH: "คุณกัลยา ก.", CustomerNameEN: "K. Kanlaya K.", AddressTH: "พร้อมพงษ์", AddressEN: "Phrom Phong", ItemCount: 31, OrderValue: 1872, DriverNameTH: "วินัย", DriverNameEN: "Winai", Status: "preparing", ETATime: "15:15", IsLate: false, DistanceKM: 1.8},
		{ID: "BC-26052205", StoreID: storeID, CustomerNameTH: "คุณธีรพล ต.", CustomerNameEN: "K. Teeraphol T.", AddressTH: "ทองหล่อ ซ.8", AddressEN: "Thonglor Soi 8", ItemCount: 6, OrderValue: 320, DriverNameTH: "สมชาย", DriverNameEN: "Somchai", Status: "delivered", ETATime: "13:50", IsLate: false, DistanceKM: 1.0},
		{ID: "BC-26052206", StoreID: storeID, CustomerNameTH: "คุณวรินทร ก.", CustomerNameEN: "K. Warintorn K.", AddressTH: "เอกมัย ซ.4", AddressEN: "Ekkamai Soi 4", ItemCount: 12, OrderValue: 654, DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong", Status: "delivered", ETATime: "13:20", IsLate: false, DistanceKM: 2.0},
		{ID: "BC-26052207", StoreID: storeID, CustomerNameTH: "คุณพิชชา ม.", CustomerNameEN: "K. Phichcha M.", AddressTH: "ทองหล่อ ซ.23", AddressEN: "Thonglor Soi 23", ItemCount: 4, OrderValue: 198, DriverNameTH: "วินัย", DriverNameEN: "Winai", Status: "delivered", ETATime: "12:45", IsLate: false, DistanceKM: 1.5},
	}
	if err := tx.Clauses(upsertByColumns([]string{"id"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert deliveries: %w", err)
	}

	return nil
}

func seedSuggestions(tx *gorm.DB, storeID int64, _ time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.Suggestion{
		{ID: "p1", StoreID: storeID, Kind: "promo", Icon: "flame", TitleTH: "ลด 30% ขนมปังฟาร์มเฮ้าส์ใกล้หมดอายุ", TitleEN: "30% off Farmhouse bread near expiry", DescriptionTH: "มีสินค้า 11 ชิ้นที่หมดอายุภายใน 2 วัน ลดราคา 30% ภายในวันนี้คาดว่าจะลดมูลค่าสูญเสียได้ ฿245", DescriptionEN: "11 units expire in 2 days. A 30% same-day markdown would recover an estimated ฿245", UpsideValue: 245, Confidence: 0.92, DurationTH: "1 วัน", DurationEN: "1 day", TargetTH: "ลูกค้าทุกคน", TargetEN: "All customers", Type: "markdown"},
		{ID: "p2", StoreID: storeID, Kind: "promo", Icon: "gift", TitleTH: "ซื้อ 2 แถม 1 — บะหมี่กึ่งสำเร็จรูป", TitleEN: "Buy 2 get 1 free — instant noodles", DescriptionTH: "ยอดขายบะหมี่ MAMA เพิ่ม 23% ในสัปดาห์นี้ จับคู่กับสินค้าเสริมเพื่อเพิ่มขนาดบิล", DescriptionEN: "MAMA noodle sales up 23% this week. Bundle to lift basket size", UpsideValue: 1840, Confidence: 0.78, DurationTH: "1 สัปดาห์", DurationEN: "1 week", TargetTH: "ลูกค้าประจำ", TargetEN: "Repeat customers", Type: "bundle"},
		{ID: "p3", StoreID: storeID, Kind: "promo", Icon: "trending", TitleTH: "Happy Hour 17:00 - 19:00 เครื่องดื่มเย็น", TitleEN: "Happy hour 17:00 - 19:00, cold drinks", DescriptionTH: "ช่วง 17:00-19:00 มียอดขายเครื่องดื่มต่ำกว่าค่าเฉลี่ย 18% ลด 10% เพื่อกระตุ้นทราฟฟิก", DescriptionEN: "Cold drink sales 18% below average in 17:00-19:00 window. A 10% promo could lift footfall", UpsideValue: 980, Confidence: 0.65, DurationTH: "2 สัปดาห์", DurationEN: "2 weeks", TargetTH: "ลูกค้าหลังเลิกงาน", TargetEN: "After-work crowd", Type: "discount"},
		{ID: "e1", StoreID: storeID, Kind: "event", Icon: "calendar", TitleTH: "เทศกาลวิสาขบูชา 31 พ.ค.", TitleEN: "Visakha Bucha · May 31", DescriptionTH: "วันหยุดนักขัตฤกษ์ ลูกค้าซื้อของไหว้พระและของแห้งเพิ่มขึ้น แนะนำให้สต็อกเทียน-ดอกไม้-น้ำดื่ม", DescriptionEN: "Public holiday. Customers stock up on offerings and dry goods — boost candles, flowers, water", UpsideValue: 8600, Confidence: 0.88, DurationTH: "5 วัน", DurationEN: "5 days", TargetTH: "ครอบครัว", TargetEN: "Families", Type: "event"},
		{ID: "e2", StoreID: storeID, Kind: "event", Icon: "gift", TitleTH: "วันแม่ 12 ส.ค. — กระเช้าของขวัญ", TitleEN: "Mother's Day · Aug 12 — gift baskets", DescriptionTH: "เริ่มเตรียมการแสดงสินค้ากระเช้าและของขวัญ 14 วันก่อนวันแม่ ดึงดูดลูกค้าที่กำลังมองหา", DescriptionEN: "Start gift-basket merchandising 14 days ahead to capture early shoppers", UpsideValue: 14200, Confidence: 0.81, DurationTH: "3 สัปดาห์", DurationEN: "3 weeks", TargetTH: "ลูกค้าที่ต้องการของขวัญ", TargetEN: "Gift shoppers", Type: "event"},
		{ID: "e3", StoreID: storeID, Kind: "event", Icon: "sparkle", TitleTH: "เปิดเทอม 16 พ.ค. — ของใช้นักเรียน", TitleEN: "Back to school · May 16", DescriptionTH: "ครัวเรือนใกล้สาขามีเด็กวัยเรียน ~38% เสนอจัดมุมเครื่องเขียน-นม-ขนมตอนเช้า", DescriptionEN: "~38% of households nearby have school-age kids. Set up a stationery + breakfast bundle corner", UpsideValue: 6400, Confidence: 0.74, DurationTH: "4 สัปดาห์", DurationEN: "4 weeks", TargetTH: "ผู้ปกครอง", TargetEN: "Parents", Type: "event"},
	}
	if err := tx.Clauses(upsertByColumns([]string{"id"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert suggestions: %w", err)
	}

	return nil
}

func upsertByColumns(columns []string) clause.OnConflict {
	conflictColumns := make([]clause.Column, 0, len(columns))
	for _, column := range columns {
		conflictColumns = append(conflictColumns, clause.Column{Name: column})
	}

	return clause.OnConflict{
		Columns:   conflictColumns,
		UpdateAll: true,
	}
}
