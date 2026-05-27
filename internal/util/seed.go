package util

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type mockStoreSeed struct {
	store            model.Store
	deliveryPrefix   string
	suggestionPrefix string
	salesScale       float64
}

// SeedMockData loads dashboard mock data using idempotent upserts.
func SeedMockData(ctx context.Context, database *gorm.DB) error {
	return database.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
		storeSeeds := []mockStoreSeed{
			{
				store: model.Store{
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
				},
				deliveryPrefix:   "BC-260522",
				suggestionPrefix: "",
				salesScale:       1,
			},
			{
				store: model.Store{
					Code:            "MBC-0178",
					NameTH:          "มินิ บิ๊กซี สาขาอารีย์ ซอย 1",
					NameEN:          "Mini BigC · Ari Soi 1",
					ShortNameTH:     "อารีย์ ซ.1",
					ShortNameEN:     "Ari 1",
					AddressTH:       "36/2 ซ.อารีย์ 1, แขวงพญาไท, พญาไท, กรุงเทพฯ 10400",
					AddressEN:       "36/2 Ari Soi 1, Phaya Thai, Bangkok 10400",
					ManagerNameTH:   "ศิริพร วงศ์วัฒนา",
					ManagerNameEN:   "Siriporn Wongwattana",
					ManagerInitials: "SW",
					StaffNameTH:     "กิตติพงศ์ แสงชัย",
					StaffNameEN:     "Kittipong Saengchai",
					StaffInitials:   "KS",
				},
				deliveryPrefix:   "BC-260523",
				suggestionPrefix: "ari-",
				salesScale:       0.82,
			},
		}

		for _, seed := range storeSeeds {
			store := seed.store
			if err := tx.Clauses(upsertByColumns([]string{"code"})).Create(&store).Error; err != nil {
				return fmt.Errorf("upsert store %s: %w", seed.store.Code, err)
			}
			if err := tx.Where("code = ?", seed.store.Code).First(&store).Error; err != nil {
				return fmt.Errorf("query seeded store %s: %w", seed.store.Code, err)
			}

			seeders := []func(*gorm.DB, mockStoreSeed, int64, time.Time, map[string]int64, map[string]int64) error{
				seedSales,
				seedCategorySales,
				seedPaymentMix,
				seedTopProducts,
				seedInventory,
				seedDeliveries,
				seedSuggestions,
			}
			for _, seeder := range seeders {
				if err := seeder(tx, seed, store.ID, today, categoryIDs, paymentMethodIDs); err != nil {
					return err
				}
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
		// Existing / core products
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
		{SKU: "FB-1107", NameTH: "ข้าวกล่องหมูทอดกระเทียม", NameEN: "Garlic pork ready meal", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FB-3201", NameTH: "แซนด์วิชทูน่าพร้อมทาน", NameEN: "Ready-to-eat tuna sandwich", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FB-7730", NameTH: "ซาลาเปาหมูแดง", NameEN: "BBQ pork steamed bun", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FB-8842", NameTH: "ขนมจีบหมูพร้อมทาน", NameEN: "Ready-to-eat pork shumai", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "PC-0150", NameTH: "ถ่านอัลคาไลน์ AAA 4 ก้อน", NameEN: "AAA alkaline batteries × 4", CategoryID: categoryIDs["Household Items"]},
		{SKU: "PC-2219", NameTH: "ลูกอมฮอลล์ รสน้ำผึ้งมะนาว", NameEN: "Halls honey lemon candy", CategoryID: categoryIDs["Snacks"]},
		{SKU: "PC-4502", NameTH: "หมากฝรั่งเดนทีน ไวท์", NameEN: "Dentyne White gum", CategoryID: categoryIDs["Snacks"]},
		{SKU: "PC-7811", NameTH: "ทิชชู่เปียกพกพา 20 แผ่น", NameEN: "Pocket wet wipes 20 sheets", CategoryID: categoryIDs["Personal"]},
		{SKU: "HH-0091", NameTH: "น้ำยาล้างห้องน้ำ วิคตอเรีย", NameEN: "Bathroom cleaner", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-4308", NameTH: "น้ำยาซักผ้าชนิดน้ำ 900 มล.", NameEN: "Liquid laundry detergent 900ml", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7204", NameTH: "แผ่นทำความสะอาดพื้น", NameEN: "Floor cleaning sheets", CategoryID: categoryIDs["Household Items"]},
		{SKU: "PB-1008", NameTH: "โลชั่นวาสลีน 250 มล.", NameEN: "Vaseline lotion 250ml", CategoryID: categoryIDs["Personal Care"]},
		{SKU: "PB-2113", NameTH: "ครีมอาบน้ำโดฟ 450 มล.", NameEN: "Dove body wash 450ml", CategoryID: categoryIDs["Personal Care"]},
		{SKU: "PB-3902", NameTH: "แป้งเด็กแคร์ 180 ก.", NameEN: "Care baby powder 180g", CategoryID: categoryIDs["Personal Care"]},
		{SKU: "FR-0110", NameTH: "ข้าวเหนียวหมูปิ้งพร้อมทาน", NameEN: "Grilled pork sticky rice", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FR-0298", NameTH: "สลัดอกไก่พร้อมทาน", NameEN: "Chicken breast salad", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "FR-3007", NameTH: "ครัวซองต์แฮมชีส", NameEN: "Ham cheese croissant", CategoryID: categoryIDs["Food & Beverage"]},
		{SKU: "DR-0044", NameTH: "น้ำดื่มคริสตัล 1.5 ลิตร", NameEN: "Crystal water 1.5L", CategoryID: categoryIDs["Beverage"]},
		{SKU: "DR-0188", NameTH: "กาแฟเย็นพร้อมดื่ม 220 มล.", NameEN: "Ready-to-drink iced coffee 220ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "DR-7120", NameTH: "ชาไทยพร้อมดื่ม 250 มล.", NameEN: "Ready-to-drink Thai tea 250ml", CategoryID: categoryIDs["Beverage"]},

		// Food
		{SKU: "FB-3301", NameTH: "มาม่า รสต้มยำกุ้ง 55 ก.", NameEN: "MAMA Tom Yum Kung 55g", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3302", NameTH: "ไวไว รสดั้งเดิม 60 ก.", NameEN: "Wai Wai Original 60g", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3303", NameTH: "โจ๊กคัพ รสหมู 45 ก.", NameEN: "Instant rice porridge pork cup 45g", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3304", NameTH: "ปลากระป๋อง สามแม่ครัว", NameEN: "Three Lady Cooks canned fish", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3305", NameTH: "ข้าวหอมมะลิ 5 กก.", NameEN: "Jasmine rice 5kg", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3306", NameTH: "น้ำปลาแท้ ทิพรส 700 มล.", NameEN: "Tiparos fish sauce 700ml", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3307", NameTH: "น้ำตาลทรายขาว 1 กก.", NameEN: "White sugar 1kg", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3308", NameTH: "น้ำมันพืช มรกต 1 ลิตร", NameEN: "Morakot vegetable oil 1L", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3309", NameTH: "ซอสหอยนางรม เด็กสมบูรณ์", NameEN: "Healthy Boy oyster sauce", CategoryID: categoryIDs["Food"]},
		{SKU: "FB-3310", NameTH: "บะหมี่คัพ นิชชิน รสซีฟู้ด", NameEN: "Nissin seafood cup noodles", CategoryID: categoryIDs["Food"]},

		// Beverage
		{SKU: "BV-1001", NameTH: "โค้ก 325 มล.", NameEN: "Coca-Cola 325ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1002", NameTH: "เป๊ปซี่ 345 มล.", NameEN: "Pepsi 345ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1003", NameTH: "สปอนเซอร์ 250 มล.", NameEN: "Sponsor energy drink 250ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1004", NameTH: "กาแฟกระป๋อง เบอร์ดี้ โรบัสต้า", NameEN: "Birdy Robusta canned coffee", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1005", NameTH: "น้ำส้มสแปลช 250 มล.", NameEN: "Splash orange drink 250ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1006", NameTH: "ชาเขียวโออิชิ น้ำผึ้งมะนาว", NameEN: "Oishi honey lemon green tea", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1007", NameTH: "น้ำดื่มสิงห์ 1.5 ลิตร", NameEN: "Singha drinking water 1.5L", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1008", NameTH: "เอ็ม-150 150 มล.", NameEN: "M-150 energy drink 150ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1009", NameTH: "นมถั่วเหลือง แลคตาซอย 300 มล.", NameEN: "Lactasoy soy milk 300ml", CategoryID: categoryIDs["Beverage"]},
		{SKU: "BV-1010", NameTH: "น้ำแร่ มิเนเร่ 600 มล.", NameEN: "Minere mineral water 600ml", CategoryID: categoryIDs["Beverage"]},

		// Snacks
		{SKU: "SN-2001", NameTH: "เทสโต รสดั้งเดิม 75 ก.", NameEN: "Tasto Original 75g", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2002", NameTH: "ป๊อกกี้ ช็อกโกแลต", NameEN: "Pocky Chocolate", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2003", NameTH: "โอรีโอ วานิลลา", NameEN: "Oreo Vanilla", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2004", NameTH: "เบนโตะ ปลาหมึกอบ", NameEN: "Bento squid snack", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2005", NameTH: "คิทแคท มินิ", NameEN: "KitKat Mini", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2006", NameTH: "ฮานามิ ข้าวเกรียบกุ้ง", NameEN: "Hanami shrimp crackers", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2007", NameTH: "คารามูโจ้ รสเผ็ด", NameEN: "Karamucho spicy chips", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2008", NameTH: "สาหร่ายเถ้าแก่น้อย", NameEN: "Tao Kae Noi seaweed", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2009", NameTH: "ถั่วโก๋แก่ รสกะทิ", NameEN: "Koh-Kae coconut peanuts", CategoryID: categoryIDs["Snacks"]},
		{SKU: "SN-2010", NameTH: "เยลลี่ปีโป้", NameEN: "Pipo jelly", CategoryID: categoryIDs["Snacks"]},

		// Chilled
		{SKU: "CH-4001", NameTH: "ไข่ไก่ เบอร์ 2 แพ็ก 10 ฟอง", NameEN: "Chicken eggs size 2 × 10", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4002", NameTH: "นมเมจิ รสจืด 450 มล.", NameEN: "Meiji plain milk 450ml", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4003", NameTH: "แฮมหมู CP 200 ก.", NameEN: "CP pork ham 200g", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4004", NameTH: "ชีสแผ่น 12 แผ่น", NameEN: "Cheese slices × 12", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4005", NameTH: "เกี๊ยวหมู CP แช่เย็น", NameEN: "CP chilled pork wonton", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4006", NameTH: "สลัดผักพร้อมทาน", NameEN: "Ready-to-eat salad", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4007", NameTH: "โยเกิร์ตเมจิ รสสตรอว์เบอร์รี", NameEN: "Meiji strawberry yoghurt", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4008", NameTH: "นมเปรี้ยวดัชมิลล์", NameEN: "Dutch Mill drinking yoghurt", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4009", NameTH: "แซนด์วิชแฮมชีส", NameEN: "Ham cheese sandwich", CategoryID: categoryIDs["Chilled"]},
		{SKU: "CH-4010", NameTH: "ข้าวกล่องกะเพราไก่", NameEN: "Chicken basil ready meal", CategoryID: categoryIDs["Chilled"]},

		// Frozen
		{SKU: "FR-5001", NameTH: "เกี๊ยวซ่าแช่แข็ง", NameEN: "Frozen gyoza", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5002", NameTH: "ข้าวผัดกุ้งแช่แข็ง", NameEN: "Frozen shrimp fried rice", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5003", NameTH: "ไก่ทอดแช่แข็ง", NameEN: "Frozen fried chicken", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5004", NameTH: "พิซซ่าแช่แข็ง", NameEN: "Frozen pizza", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5005", NameTH: "ไอศกรีมวอลล์ คัพ", NameEN: "Wall's ice cream cup", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5006", NameTH: "หมูบดแช่แข็ง", NameEN: "Frozen minced pork", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5007", NameTH: "นักเก็ตไก่แช่แข็ง", NameEN: "Frozen chicken nuggets", CategoryID: categoryIDs["Frozen"]},
		{SKU: "FR-5008", NameTH: "ลูกชิ้นปลาแช่แข็ง", NameEN: "Frozen fish balls", CategoryID: categoryIDs["Frozen"]},

		// Personal
		{SKU: "PC-6001", NameTH: "แชมพูซันซิล 320 มล.", NameEN: "Sunsilk shampoo 320ml", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6002", NameTH: "ยาสีฟันคอลเกต 150 ก.", NameEN: "Colgate toothpaste 150g", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6003", NameTH: "สบู่ลักส์ 4 ก้อน", NameEN: "Lux soap × 4", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6004", NameTH: "โฟมล้างหน้า นีเวีย", NameEN: "Nivea facial foam", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6005", NameTH: "โรลออนนีเวีย", NameEN: "Nivea roll-on", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6006", NameTH: "แปรงสีฟันออรัลบี", NameEN: "Oral-B toothbrush", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6007", NameTH: "ครีมกันแดดการ์นิเย่", NameEN: "Garnier sunscreen", CategoryID: categoryIDs["Personal"]},
		{SKU: "PC-6008", NameTH: "ทิชชู่เปียก 80 แผ่น", NameEN: "Wet wipes 80 sheets", CategoryID: categoryIDs["Personal"]},

		// Household
		{SKU: "HH-7001", NameTH: "น้ำยาล้างจาน ซันไลต์ 550 มล.", NameEN: "Sunlight dishwashing liquid 550ml", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7002", NameTH: "น้ำยาปรับผ้านุ่ม ดาวน์นี่ 500 มล.", NameEN: "Downy fabric softener 500ml", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7003", NameTH: "ถุงขยะดำ 24x28 นิ้ว", NameEN: "Black garbage bags 24x28 inch", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7004", NameTH: "ฟองน้ำล้างจาน 3 ชิ้น", NameEN: "Dish sponge × 3", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7005", NameTH: "น้ำยาถูพื้น มาจิคลีน", NameEN: "Magiclean floor cleaner", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7006", NameTH: "สเปรย์กำจัดยุง", NameEN: "Mosquito spray", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7007", NameTH: "ถ่านอัลคาไลน์ AA 4 ก้อน", NameEN: "AA alkaline batteries × 4", CategoryID: categoryIDs["Household Items"]},
		{SKU: "HH-7008", NameTH: "กระดาษเช็ดครัว", NameEN: "Kitchen paper towel", CategoryID: categoryIDs["Household Items"]},

		// Produce
		{SKU: "PD-8001", NameTH: "กล้วยหอม 1 หวี", NameEN: "Banana bunch", CategoryID: categoryIDs["Produce"]},
		{SKU: "PD-8002", NameTH: "แอปเปิลฟูจิ 4 ลูก", NameEN: "Fuji apples × 4", CategoryID: categoryIDs["Produce"]},
		{SKU: "PD-8003", NameTH: "มะเขือเทศราชินี 250 ก.", NameEN: "Cherry tomatoes 250g", CategoryID: categoryIDs["Produce"]},
		{SKU: "PD-8004", NameTH: "ผักสลัดรวม", NameEN: "Mixed salad vegetables", CategoryID: categoryIDs["Produce"]},
		{SKU: "PD-8005", NameTH: "แตงโมหั่นพร้อมทาน", NameEN: "Ready-to-eat watermelon", CategoryID: categoryIDs["Produce"]},
	}

	if err := tx.Clauses(upsertByColumns([]string{"sku"})).Create(&products).Error; err != nil {
		return fmt.Errorf("upsert products: %w", err)
	}

	return nil
}

func seedSales(tx *gorm.DB, seed mockStoreSeed, storeID int64, today time.Time, _ map[string]int64, _ map[string]int64) error {
	// 24-hour sales pattern.
	// Useful for dashboard line charts.
	hourlyBase := []float64{
		900, 700, 500, 450, 600, 1200,
		3200, 4100, 5800, 8400, 11200, 9800,
		12400, 15600, 14200, 11800, 9400, 13200,
		18900, 21400, 19200, 14800, 7600, 3100,
	}

	hourly := make([]model.SalesHourly, 0, len(hourlyBase))

	for hour, value := range hourlyBase {
		var multiplier float64 = 1

		// Morning rush
		if hour >= 7 && hour <= 9 {
			multiplier = 1.12
		}

		// Lunch period
		if hour >= 11 && hour <= 13 {
			multiplier = 1.18
		}

		// Evening peak
		if hour >= 17 && hour <= 20 {
			multiplier = 1.28
		}

		// Late night lower traffic
		if hour >= 0 && hour <= 5 {
			multiplier = 0.72
		}

		currentValue := value * multiplier
		comparisonValue := currentValue * 0.91

		hourly = append(hourly, model.SalesHourly{
			StoreID:              storeID,
			SaleDate:             today,
			Hour:                 hour,
			SalesValue:           scaledCurrency(currentValue, seed.salesScale),
			ComparisonSalesValue: scaledCurrency(comparisonValue, seed.salesScale),
		})
	}

	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sale_date", "hour"})).Create(&hourly).Error; err != nil {
		return fmt.Errorf("upsert hourly sales: %w", err)
	}

	// 30-day sales data.
	// Useful for trend chart, weekly comparison, and dashboard KPI.
	daily := make([]model.SalesDaily, 0, 30)

	for i := 29; i >= 0; i-- {
		date := today.AddDate(0, 0, -i)

		baseSales := 178000.0

		weekday := date.Weekday()

		switch weekday {
		case time.Friday:
			baseSales *= 1.16
		case time.Saturday:
			baseSales *= 1.32
		case time.Sunday:
			baseSales *= 1.22
		case time.Monday:
			baseSales *= 0.91
		case time.Tuesday:
			baseSales *= 0.96
		case time.Wednesday:
			baseSales *= 1.00
		case time.Thursday:
			baseSales *= 1.04
		}

		// Add simple repeating fluctuation.
		// This makes mock data look less flat.
		fluctuation := 1 + (float64((30-i)%7)-3)*0.025

		currentSales := baseSales * fluctuation

		// Previous comparison period.
		comparisonSales := currentSales * 0.93

		// Simulate campaign day.
		if i == 5 || i == 12 || i == 21 {
			currentSales *= 1.18
		}

		// Simulate rainy / low traffic day.
		if i == 8 || i == 17 {
			currentSales *= 0.82
		}

		daily = append(daily, model.SalesDaily{
			StoreID:              storeID,
			SaleDate:             date,
			SalesValue:           scaledCurrency(currentSales, seed.salesScale),
			ComparisonSalesValue: scaledCurrency(comparisonSales, seed.salesScale),
		})
	}

	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sale_date"})).Create(&daily).Error; err != nil {
		return fmt.Errorf("upsert daily sales: %w", err)
	}

	// 12-month sales data.
	// Useful for yearly overview chart.
	monthlyBase := []float64{
		5420000,
		5180000,
		5960000,
		6180000,
		6740000,
		6920000,
		7180000,
		7040000,
		6880000,
		7320000,
		7860000,
		8140000,
	}

	monthly := make([]model.SalesMonthly, 0, len(monthlyBase))

	for index, value := range monthlyBase {
		month := index + 1

		// Add seasonal pattern.
		seasonalValue := value

		// Songkran / summer period
		if month == 4 {
			seasonalValue *= 1.12
		}

		// Back to school
		if month == 5 {
			seasonalValue *= 1.06
		}

		// Year-end shopping
		if month == 12 {
			seasonalValue *= 1.18
		}

		monthly = append(monthly, model.SalesMonthly{
			StoreID:    storeID,
			Year:       today.Year(),
			Month:      month,
			SalesValue: scaledCurrency(seasonalValue, seed.salesScale),
		})
	}

	if err := tx.Clauses(upsertByColumns([]string{"store_id", "year", "month"})).Create(&monthly).Error; err != nil {
		return fmt.Errorf("upsert monthly sales: %w", err)
	}

	return nil
}

func seedCategorySales(tx *gorm.DB, seed mockStoreSeed, storeID int64, today time.Time, categoryIDs map[string]int64, _ map[string]int64) error {
	rows := []model.CategorySalesEntity{
		{StoreID: storeID, CategoryID: categoryIDs["Food & Beverage"], SalesDate: today, SalesValue: scaledCurrency(142800, seed.salesScale), Share: 0.38, TrendPercent: 6.2},
		{StoreID: storeID, CategoryID: categoryIDs["Household"], SalesDate: today, SalesValue: scaledCurrency(78400, seed.salesScale), Share: 0.21, TrendPercent: -2.1},
		{StoreID: storeID, CategoryID: categoryIDs["Personal Care"], SalesDate: today, SalesValue: scaledCurrency(56200, seed.salesScale), Share: 0.15, TrendPercent: 4.8},
		{StoreID: storeID, CategoryID: categoryIDs["Snacks"], SalesDate: today, SalesValue: scaledCurrency(49100, seed.salesScale), Share: 0.13, TrendPercent: 11.3},
		{StoreID: storeID, CategoryID: categoryIDs["Frozen"], SalesDate: today, SalesValue: scaledCurrency(28400, seed.salesScale), Share: 0.075, TrendPercent: -1.4},
		{StoreID: storeID, CategoryID: categoryIDs["Other"], SalesDate: today, SalesValue: scaledCurrency(21100, seed.salesScale), Share: 0.055, TrendPercent: 2.0},
	}
	if err := tx.Clauses(upsertByColumns([]string{"store_id", "category_id", "sales_date"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert category sales: %w", err)
	}

	return nil
}

func seedPaymentMix(tx *gorm.DB, _ mockStoreSeed, storeID int64, today time.Time, _ map[string]int64, paymentMethodIDs map[string]int64) error {
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

func seedTopProducts(tx *gorm.DB, seed mockStoreSeed, storeID int64, today time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.TopProductEntity{
		// Food & Beverage
		{StoreID: storeID, SKU: "FB-0102", SalesDate: today, SoldQuantity: scaledQuantity(184, seed.salesScale), SalesValue: scaledCurrency(9568, seed.salesScale), TrendPercent: 12},
		{StoreID: storeID, SKU: "FB-2284", SalesDate: today, SoldQuantity: scaledQuantity(162, seed.salesScale), SalesValue: scaledCurrency(4374, seed.salesScale), TrendPercent: 8},
		{StoreID: storeID, SKU: "FB-0411", SalesDate: today, SoldQuantity: scaledQuantity(148, seed.salesScale), SalesValue: scaledCurrency(2516, seed.salesScale), TrendPercent: 5},
		{StoreID: storeID, SKU: "FB-0099", SalesDate: today, SoldQuantity: scaledQuantity(121, seed.salesScale), SalesValue: scaledCurrency(8470, seed.salesScale), TrendPercent: -2},
		{StoreID: storeID, SKU: "FB-1107", SalesDate: today, SoldQuantity: scaledQuantity(139, seed.salesScale), SalesValue: scaledCurrency(4865, seed.salesScale), TrendPercent: 9},
		{StoreID: storeID, SKU: "FB-3201", SalesDate: today, SoldQuantity: scaledQuantity(116, seed.salesScale), SalesValue: scaledCurrency(5800, seed.salesScale), TrendPercent: 4},
		{StoreID: storeID, SKU: "FB-7730", SalesDate: today, SoldQuantity: scaledQuantity(104, seed.salesScale), SalesValue: scaledCurrency(3120, seed.salesScale), TrendPercent: -6},
		{StoreID: storeID, SKU: "FB-8842", SalesDate: today, SoldQuantity: scaledQuantity(98, seed.salesScale), SalesValue: scaledCurrency(4410, seed.salesScale), TrendPercent: 7},

		// Packaged / Consumer Products
		{StoreID: storeID, SKU: "PC-0331", SalesDate: today, SoldQuantity: scaledQuantity(96, seed.salesScale), SalesValue: scaledCurrency(11520, seed.salesScale), TrendPercent: 18},
		{StoreID: storeID, SKU: "PC-0150", SalesDate: today, SoldQuantity: scaledQuantity(91, seed.salesScale), SalesValue: scaledCurrency(4095, seed.salesScale), TrendPercent: 6},
		{StoreID: storeID, SKU: "PC-2219", SalesDate: today, SoldQuantity: scaledQuantity(84, seed.salesScale), SalesValue: scaledCurrency(7560, seed.salesScale), TrendPercent: 11},
		{StoreID: storeID, SKU: "PC-4502", SalesDate: today, SoldQuantity: scaledQuantity(76, seed.salesScale), SalesValue: scaledCurrency(6080, seed.salesScale), TrendPercent: -3},
		{StoreID: storeID, SKU: "PC-7811", SalesDate: today, SoldQuantity: scaledQuantity(69, seed.salesScale), SalesValue: scaledCurrency(5520, seed.salesScale), TrendPercent: 2},

		// Household
		{StoreID: storeID, SKU: "HH-1820", SalesDate: today, SoldQuantity: scaledQuantity(88, seed.salesScale), SalesValue: scaledCurrency(7392, seed.salesScale), TrendPercent: 3},
		{StoreID: storeID, SKU: "HH-0091", SalesDate: today, SoldQuantity: scaledQuantity(73, seed.salesScale), SalesValue: scaledCurrency(6570, seed.salesScale), TrendPercent: 5},
		{StoreID: storeID, SKU: "HH-4308", SalesDate: today, SoldQuantity: scaledQuantity(61, seed.salesScale), SalesValue: scaledCurrency(5490, seed.salesScale), TrendPercent: -4},
		{StoreID: storeID, SKU: "HH-7204", SalesDate: today, SoldQuantity: scaledQuantity(54, seed.salesScale), SalesValue: scaledCurrency(4860, seed.salesScale), TrendPercent: 1},

		// Personal Care
		{StoreID: storeID, SKU: "PB-1008", SalesDate: today, SoldQuantity: scaledQuantity(82, seed.salesScale), SalesValue: scaledCurrency(9020, seed.salesScale), TrendPercent: 14},
		{StoreID: storeID, SKU: "PB-2113", SalesDate: today, SoldQuantity: scaledQuantity(74, seed.salesScale), SalesValue: scaledCurrency(7770, seed.salesScale), TrendPercent: 10},
		{StoreID: storeID, SKU: "PB-3902", SalesDate: today, SoldQuantity: scaledQuantity(58, seed.salesScale), SalesValue: scaledCurrency(6670, seed.salesScale), TrendPercent: -1},

		// Fresh / Ready-to-eat
		{StoreID: storeID, SKU: "FR-0110", SalesDate: today, SoldQuantity: scaledQuantity(128, seed.salesScale), SalesValue: scaledCurrency(3840, seed.salesScale), TrendPercent: 15},
		{StoreID: storeID, SKU: "FR-0298", SalesDate: today, SoldQuantity: scaledQuantity(112, seed.salesScale), SalesValue: scaledCurrency(5040, seed.salesScale), TrendPercent: 9},
		{StoreID: storeID, SKU: "FR-3007", SalesDate: today, SoldQuantity: scaledQuantity(93, seed.salesScale), SalesValue: scaledCurrency(4185, seed.salesScale), TrendPercent: -5},

		// Drinks
		{StoreID: storeID, SKU: "DR-0044", SalesDate: today, SoldQuantity: scaledQuantity(157, seed.salesScale), SalesValue: scaledCurrency(4710, seed.salesScale), TrendPercent: 13},
		{StoreID: storeID, SKU: "DR-0188", SalesDate: today, SoldQuantity: scaledQuantity(143, seed.salesScale), SalesValue: scaledCurrency(5005, seed.salesScale), TrendPercent: 6},
		{StoreID: storeID, SKU: "DR-7120", SalesDate: today, SoldQuantity: scaledQuantity(97, seed.salesScale), SalesValue: scaledCurrency(3395, seed.salesScale), TrendPercent: -8},
	}

	if err := tx.Clauses(upsertByColumns([]string{"store_id", "sku", "sales_date"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert top products: %w", err)
	}

	return nil
}

func seedInventory(tx *gorm.DB, _ mockStoreSeed, storeID int64, today time.Time, categoryIDs map[string]int64, _ map[string]int64) error {
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

func seedDeliveries(tx *gorm.DB, seed mockStoreSeed, storeID int64, _ time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.Delivery{
		{
			ID: seed.deliveryPrefix + "01", StoreID: storeID,
			CustomerNameTH: "คุณณัฐกานต์ ม.", CustomerNameEN: "K. Natthakarn M.",
			AddressTH: "ทองหล่อ ซ.10", AddressEN: "Thonglor Soi 10",
			ItemCount: 8, OrderValue: scaledCurrency(487, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "enRoute", ETATime: "14:25", IsLate: false, DistanceKM: 1.2,
		},
		{
			ID: seed.deliveryPrefix + "02", StoreID: storeID,
			CustomerNameTH: "คุณปานทิพย์ ส.", CustomerNameEN: "K. Panthip S.",
			AddressTH: "เอกมัย ซ.12", AddressEN: "Ekkamai Soi 12",
			ItemCount: 22, OrderValue: scaledCurrency(1284, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "enRoute", ETATime: "14:35", IsLate: true, DistanceKM: 2.4,
		},
		{
			ID: seed.deliveryPrefix + "03", StoreID: storeID,
			CustomerNameTH: "คุณภาสกร ว.", CustomerNameEN: "K. Phasakorn W.",
			AddressTH: "ทองหล่อ ซ.5", AddressEN: "Thonglor Soi 5",
			ItemCount: 14, OrderValue: scaledCurrency(892, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "preparing", ETATime: "15:00", IsLate: false, DistanceKM: 0.8,
		},
		{
			ID: seed.deliveryPrefix + "04", StoreID: storeID,
			CustomerNameTH: "คุณกัลยา ก.", CustomerNameEN: "K. Kanlaya K.",
			AddressTH: "พร้อมพงษ์", AddressEN: "Phrom Phong",
			ItemCount: 31, OrderValue: scaledCurrency(1872, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "preparing", ETATime: "15:15", IsLate: false, DistanceKM: 1.8,
		},
		{
			ID: seed.deliveryPrefix + "05", StoreID: storeID,
			CustomerNameTH: "คุณธีรพล ต.", CustomerNameEN: "K. Teeraphol T.",
			AddressTH: "ทองหล่อ ซ.8", AddressEN: "Thonglor Soi 8",
			ItemCount: 6, OrderValue: scaledCurrency(320, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "delivered", ETATime: "13:50", IsLate: false, DistanceKM: 1.0,
		},
		{
			ID: seed.deliveryPrefix + "06", StoreID: storeID,
			CustomerNameTH: "คุณวรินทร ก.", CustomerNameEN: "K. Warintorn K.",
			AddressTH: "เอกมัย ซ.4", AddressEN: "Ekkamai Soi 4",
			ItemCount: 12, OrderValue: scaledCurrency(654, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "delivered", ETATime: "13:20", IsLate: false, DistanceKM: 2.0,
		},
		{
			ID: seed.deliveryPrefix + "07", StoreID: storeID,
			CustomerNameTH: "คุณพิชชา ม.", CustomerNameEN: "K. Phichcha M.",
			AddressTH: "ทองหล่อ ซ.23", AddressEN: "Thonglor Soi 23",
			ItemCount: 4, OrderValue: scaledCurrency(198, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "delivered", ETATime: "12:45", IsLate: false, DistanceKM: 1.5,
		},
		{
			ID: seed.deliveryPrefix + "08", StoreID: storeID,
			CustomerNameTH: "คุณอรทัย พ.", CustomerNameEN: "K. Orathai P.",
			AddressTH: "คอนโดใกล้สถานี BTS", AddressEN: "Condo near BTS station",
			ItemCount: 19, OrderValue: scaledCurrency(1045, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "enRoute", ETATime: "15:45", IsLate: true, DistanceKM: 3.1,
		},
		{
			ID: seed.deliveryPrefix + "09", StoreID: storeID,
			CustomerNameTH: "คุณธนกร จ.", CustomerNameEN: "K. Thanakorn J.",
			AddressTH: "อาคารสำนักงาน", AddressEN: "Office building",
			ItemCount: 3, OrderValue: scaledCurrency(129, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "preparing", ETATime: "14:10", IsLate: false, DistanceKM: 0.6,
		},
		{
			ID: seed.deliveryPrefix + "10", StoreID: storeID,
			CustomerNameTH: "คุณลลิตา น.", CustomerNameEN: "K. Lalita N.",
			AddressTH: "หอพักนักศึกษา", AddressEN: "Student dormitory",
			ItemCount: 16, OrderValue: scaledCurrency(739, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "enRoute", ETATime: "13:55", IsLate: true, DistanceKM: 2.7,
		},
		{
			ID: seed.deliveryPrefix + "11", StoreID: storeID,
			CustomerNameTH: "คุณเมธาวี ร.", CustomerNameEN: "K. Methawee R.",
			AddressTH: "คอนโดใกล้โรงพยาบาล", AddressEN: "Condo near hospital",
			ItemCount: 9, OrderValue: scaledCurrency(512, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "preparing", ETATime: "16:05", IsLate: false, DistanceKM: 1.7,
		},
		{
			ID: seed.deliveryPrefix + "12", StoreID: storeID,
			CustomerNameTH: "คุณวราพร ค.", CustomerNameEN: "K. Waraporn K.",
			AddressTH: "อพาร์ตเมนต์หลังตลาด", AddressEN: "Apartment behind market",
			ItemCount: 27, OrderValue: scaledCurrency(1590, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "enRoute", ETATime: "16:20", IsLate: false, DistanceKM: 2.9,
		},
		{
			ID: seed.deliveryPrefix + "13", StoreID: storeID,
			CustomerNameTH: "คุณพีรพล ย.", CustomerNameEN: "K. Peerapol Y.",
			AddressTH: "หมู่บ้านกลางเมือง", AddressEN: "Klang Muang village",
			ItemCount: 5, OrderValue: scaledCurrency(275, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "delivered", ETATime: "12:10", IsLate: false, DistanceKM: 3.4,
		},
		{
			ID: seed.deliveryPrefix + "14", StoreID: storeID,
			CustomerNameTH: "คุณชาลิสา ด.", CustomerNameEN: "K. Chalisa D.",
			AddressTH: "ตึกสำนักงานใกล้แยกไฟแดง", AddressEN: "Office near traffic junction",
			ItemCount: 18, OrderValue: scaledCurrency(965, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "enRoute", ETATime: "16:40", IsLate: true, DistanceKM: 4.2,
		},
		{
			ID: seed.deliveryPrefix + "15", StoreID: storeID,
			CustomerNameTH: "คุณณิชา ศ.", CustomerNameEN: "K. Nicha S.",
			AddressTH: "ซอยโรงเรียน", AddressEN: "School alley",
			ItemCount: 11, OrderValue: scaledCurrency(430, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "preparing", ETATime: "17:00", IsLate: false, DistanceKM: 1.1,
		},
		{
			ID: seed.deliveryPrefix + "16", StoreID: storeID,
			CustomerNameTH: "คุณกิตติชัย ล.", CustomerNameEN: "K. Kittichai L.",
			AddressTH: "บ้านพักใกล้สวนสาธารณะ", AddressEN: "House near public park",
			ItemCount: 24, OrderValue: scaledCurrency(1340, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "delivered", ETATime: "11:40", IsLate: false, DistanceKM: 2.2,
		},
		{
			ID: seed.deliveryPrefix + "17", StoreID: storeID,
			CustomerNameTH: "คุณสุรีย์พร จ.", CustomerNameEN: "K. Sureeporn J.",
			AddressTH: "คอนโดชั้น 18", AddressEN: "Condo floor 18",
			ItemCount: 7, OrderValue: scaledCurrency(388, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "preparing", ETATime: "17:20", IsLate: false, DistanceKM: 0.9,
		},
		{
			ID: seed.deliveryPrefix + "18", StoreID: storeID,
			CustomerNameTH: "คุณภัทรพล อ.", CustomerNameEN: "K. Phattharaphon A.",
			AddressTH: "โฮมออฟฟิศ", AddressEN: "Home office",
			ItemCount: 33, OrderValue: scaledCurrency(2115, seed.salesScale),
			DriverNameTH: "สมชาย", DriverNameEN: "Somchai",
			Status: "enRoute", ETATime: "17:35", IsLate: false, DistanceKM: 3.8,
		},
		{
			ID: seed.deliveryPrefix + "19", StoreID: storeID,
			CustomerNameTH: "คุณมินตรา บ.", CustomerNameEN: "K. Mintra B.",
			AddressTH: "บ้านใกล้สถานีตำรวจ", AddressEN: "House near police station",
			ItemCount: 13, OrderValue: scaledCurrency(690, seed.salesScale),
			DriverNameTH: "บุญส่ง", DriverNameEN: "Boonsong",
			Status: "delivered", ETATime: "10:55", IsLate: false, DistanceKM: 2.5,
		},
		{
			ID: seed.deliveryPrefix + "20", StoreID: storeID,
			CustomerNameTH: "คุณศุภกิจ ป.", CustomerNameEN: "K. Supakit P.",
			AddressTH: "แฟลตพนักงาน", AddressEN: "Staff apartment",
			ItemCount: 21, OrderValue: scaledCurrency(1175, seed.salesScale),
			DriverNameTH: "วินัย", DriverNameEN: "Winai",
			Status: "enRoute", ETATime: "18:00", IsLate: true, DistanceKM: 4.6,
		},
	}

	if err := tx.Clauses(upsertByColumns([]string{"id"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert deliveries: %w", err)
	}

	return nil
}

func seedSuggestions(tx *gorm.DB, seed mockStoreSeed, storeID int64, _ time.Time, _ map[string]int64, _ map[string]int64) error {
	rows := []model.Suggestion{
		{
			ID: seed.suggestionPrefix + "p1", StoreID: storeID,
			Kind: "promo", Icon: "flame",
			TitleTH:       "ลด 30% ขนมปังฟาร์มเฮ้าส์ใกล้หมดอายุ",
			TitleEN:       "30% off Farmhouse bread near expiry",
			DescriptionTH: "มีสินค้า 11 ชิ้นที่หมดอายุภายใน 2 วัน ลดราคา 30% ภายในวันนี้คาดว่าจะลดมูลค่าสูญเสียได้ ฿245",
			DescriptionEN: "11 units expire in 2 days. A 30% same-day markdown would recover an estimated ฿245",
			UpsideValue:   scaledCurrency(245, seed.salesScale),
			Confidence:    0.92,
			DurationTH:    "1 วัน",
			DurationEN:    "1 day",
			TargetTH:      "ลูกค้าทุกคน",
			TargetEN:      "All customers",
			Type:          "markdown",
		},
		{
			ID: seed.suggestionPrefix + "p2", StoreID: storeID,
			Kind: "promo", Icon: "gift",
			TitleTH:       "ซื้อ 2 แถม 1 — บะหมี่กึ่งสำเร็จรูป",
			TitleEN:       "Buy 2 get 1 free — instant noodles",
			DescriptionTH: "ยอดขายบะหมี่ MAMA เพิ่มขึ้นในสัปดาห์นี้ จับคู่กับสินค้าเสริมเพื่อเพิ่มขนาดบิล",
			DescriptionEN: "MAMA noodle sales are increasing this week. Bundle with add-on items to lift basket size",
			UpsideValue:   scaledCurrency(1840, seed.salesScale),
			Confidence:    0.78,
			DurationTH:    "1 สัปดาห์",
			DurationEN:    "1 week",
			TargetTH:      "ลูกค้าประจำ",
			TargetEN:      "Repeat customers",
			Type:          "bundle",
		},
		{
			ID: seed.suggestionPrefix + "p3", StoreID: storeID,
			Kind: "promo", Icon: "trending",
			TitleTH:       "Happy Hour 17:00 - 19:00 เครื่องดื่มเย็น",
			TitleEN:       "Happy hour 17:00 - 19:00, cold drinks",
			DescriptionTH: "ช่วง 17:00-19:00 มียอดขายเครื่องดื่มต่ำกว่าค่าเฉลี่ย ลด 10% เพื่อกระตุ้นทราฟฟิกช่วงเย็น",
			DescriptionEN: "Cold drink sales are below average between 17:00-19:00. A 10% promo could lift evening footfall",
			UpsideValue:   scaledCurrency(980, seed.salesScale),
			Confidence:    0.65,
			DurationTH:    "2 สัปดาห์",
			DurationEN:    "2 weeks",
			TargetTH:      "ลูกค้าหลังเลิกงาน",
			TargetEN:      "After-work crowd",
			Type:          "discount",
		},
		{
			ID: seed.suggestionPrefix + "e1", StoreID: storeID,
			Kind: "event", Icon: "calendar",
			TitleTH:       "เทศกาลวิสาขบูชา — เพิ่มสินค้าทำบุญ",
			TitleEN:       "Visakha Bucha — increase merit-making items",
			DescriptionTH: "ลูกค้ามักซื้อของไหว้พระ น้ำดื่ม และอาหารแห้งเพิ่มขึ้น แนะนำเพิ่มพื้นที่โชว์สินค้าใกล้ทางเข้า",
			DescriptionEN: "Customers often buy offerings, water, and dry goods. Increase display space near the entrance",
			UpsideValue:   scaledCurrency(8600, seed.salesScale),
			Confidence:    0.88,
			DurationTH:    "5 วัน",
			DurationEN:    "5 days",
			TargetTH:      "ครอบครัว",
			TargetEN:      "Families",
			Type:          "event",
		},
		{
			ID: seed.suggestionPrefix + "e2", StoreID: storeID,
			Kind: "event", Icon: "gift",
			TitleTH:       "วันแม่ — กระเช้าของขวัญและสินค้าเพื่อสุขภาพ",
			TitleEN:       "Mother's Day — gift baskets and health products",
			DescriptionTH: "เริ่มจัดมุมกระเช้าและสินค้าเพื่อสุขภาพก่อนวันแม่ 14 วัน เพื่อจับกลุ่มลูกค้าที่ซื้อของขวัญล่วงหน้า",
			DescriptionEN: "Start gift basket and wellness displays 14 days before Mother's Day to capture early gift shoppers",
			UpsideValue:   scaledCurrency(14200, seed.salesScale),
			Confidence:    0.81,
			DurationTH:    "3 สัปดาห์",
			DurationEN:    "3 weeks",
			TargetTH:      "ลูกค้าที่ต้องการของขวัญ",
			TargetEN:      "Gift shoppers",
			Type:          "event",
		},
		{
			ID: seed.suggestionPrefix + "e3", StoreID: storeID,
			Kind: "event", Icon: "sparkle",
			TitleTH:       "เปิดเทอม — มุมอาหารเช้าและของใช้นักเรียน",
			TitleEN:       "Back to school — breakfast and student supplies",
			DescriptionTH: "จัดมุมเครื่องเขียน นม ขนมปัง และอาหารเช้าพร้อมทาน ช่วยเพิ่มยอดขายช่วง 06:30-08:30",
			DescriptionEN: "Set up stationery, milk, bread, and ready-to-eat breakfast items to increase 06:30-08:30 sales",
			UpsideValue:   scaledCurrency(6400, seed.salesScale),
			Confidence:    0.74,
			DurationTH:    "4 สัปดาห์",
			DurationEN:    "4 weeks",
			TargetTH:      "ผู้ปกครองและนักเรียน",
			TargetEN:      "Parents and students",
			Type:          "event",
		},
		{
			ID: seed.suggestionPrefix + "op1", StoreID: storeID,
			Kind: "operation", Icon: "package",
			TitleTH:       "เติมสต็อกสินค้าขายดีช่วงเย็น",
			TitleEN:       "Restock evening best sellers",
			DescriptionTH: "น้ำดื่ม เครื่องดื่มเย็น และขนมขายดีช่วง 17:00-20:00 แนะนำเติมชั้นวางก่อน 16:30",
			DescriptionEN: "Water, cold drinks, and snacks sell strongly during 17:00-20:00. Refill shelves before 16:30",
			UpsideValue:   scaledCurrency(2200, seed.salesScale),
			Confidence:    0.83,
			DurationTH:    "ทุกวัน",
			DurationEN:    "Daily",
			TargetTH:      "ลูกค้าช่วงเย็น",
			TargetEN:      "Evening shoppers",
			Type:          "operation",
		},
		{
			ID: seed.suggestionPrefix + "op2", StoreID: storeID,
			Kind: "operation", Icon: "users",
			TitleTH:       "เพิ่มพนักงานแคชเชียร์ช่วง 18:00-20:00",
			TitleEN:       "Add cashier coverage from 18:00-20:00",
			DescriptionTH: "ยอดขายและจำนวนบิลสูงสุดอยู่ช่วงเย็น แนะนำเพิ่มพนักงานหน้าเคาน์เตอร์เพื่อลดเวลารอคิว",
			DescriptionEN: "Sales and bill count peak in the evening. Add cashier coverage to reduce queue time",
			UpsideValue:   scaledCurrency(1750, seed.salesScale),
			Confidence:    0.79,
			DurationTH:    "2 ชั่วโมงต่อวัน",
			DurationEN:    "2 hours per day",
			TargetTH:      "ลูกค้าหลังเลิกงาน",
			TargetEN:      "After-work customers",
			Type:          "staffing",
		},
		{
			ID: seed.suggestionPrefix + "op3", StoreID: storeID,
			Kind: "operation", Icon: "truck",
			TitleTH:       "จัดกลุ่มออเดอร์ส่งของตามโซน",
			TitleEN:       "Batch delivery orders by zone",
			DescriptionTH: "มีหลายออเดอร์ไปในพื้นที่ใกล้กัน แนะนำจัดเส้นทางตามโซนเพื่อลดเวลาขับและค่าน้ำมัน",
			DescriptionEN: "Several orders are going to nearby areas. Batch by zone to reduce travel time and fuel cost",
			UpsideValue:   scaledCurrency(1180, seed.salesScale),
			Confidence:    0.72,
			DurationTH:    "วันนี้",
			DurationEN:    "Today",
			TargetTH:      "ทีมเดลิเวอรี่",
			TargetEN:      "Delivery team",
			Type:          "delivery",
		},
		{
			ID: seed.suggestionPrefix + "inv1", StoreID: storeID,
			Kind: "inventory", Icon: "alert-triangle",
			TitleTH:       "สินค้าเครื่องดื่มใกล้หมด — ควรสั่งเพิ่ม",
			TitleEN:       "Beverage stock is low — reorder recommended",
			DescriptionTH: "น้ำดื่มและโค้กมีสต็อกต่ำกว่าระดับปลอดภัย หากไม่เติมอาจขาดสินค้าในช่วงเย็น",
			DescriptionEN: "Water and Coke stock are below safety level. Without replenishment, stockout may happen during evening peak",
			UpsideValue:   scaledCurrency(3600, seed.salesScale),
			Confidence:    0.87,
			DurationTH:    "ภายในวันนี้",
			DurationEN:    "Within today",
			TargetTH:      "ผู้จัดการสาขา",
			TargetEN:      "Store manager",
			Type:          "reorder",
		},
		{
			ID: seed.suggestionPrefix + "inv2", StoreID: storeID,
			Kind: "inventory", Icon: "snowflake",
			TitleTH:       "ตรวจอุณหภูมิสินค้าแช่เย็น",
			TitleEN:       "Check chilled product temperature",
			DescriptionTH: "สินค้าแช่เย็นมีอายุสั้นและมีความเสี่ยงสูญเสียสูง แนะนำตรวจอุณหภูมิรอบบ่ายและก่อนปิดกะ",
			DescriptionEN: "Chilled products have short shelf life and high shrink risk. Check temperature in afternoon and before shift close",
			UpsideValue:   scaledCurrency(950, seed.salesScale),
			Confidence:    0.69,
			DurationTH:    "วันนี้",
			DurationEN:    "Today",
			TargetTH:      "พนักงานกะบ่าย",
			TargetEN:      "Afternoon shift staff",
			Type:          "quality",
		},
		{
			ID: seed.suggestionPrefix + "inv3", StoreID: storeID,
			Kind: "inventory", Icon: "rotate-ccw",
			TitleTH:       "หมุนเวียนสินค้าแบบ FIFO",
			TitleEN:       "Rotate stock using FIFO",
			DescriptionTH: "สินค้ากลุ่มนม ขนมปัง และอาหารพร้อมทานควรจัดเรียงล็อตเก่าด้านหน้าเพื่อลดของหมดอายุ",
			DescriptionEN: "Milk, bread, and ready-to-eat meals should place older lots in front to reduce expiry loss",
			UpsideValue:   scaledCurrency(1320, seed.salesScale),
			Confidence:    0.84,
			DurationTH:    "ทุกกะ",
			DurationEN:    "Every shift",
			TargetTH:      "พนักงานเติมสินค้า",
			TargetEN:      "Shelf replenishment staff",
			Type:          "fifo",
		},
		{
			ID: seed.suggestionPrefix + "lp1", StoreID: storeID,
			Kind: "risk", Icon: "shield",
			TitleTH:       "ตรวจสอบสินค้าหายง่ายใกล้ทางออก",
			TitleEN:       "Check high-shrink items near exit",
			DescriptionTH: "สินค้าขนาดเล็กและราคาสูงควรย้ายไปใกล้แคชเชียร์หรือติดป้ายเตือนเพื่อลดการสูญหาย",
			DescriptionEN: "Small high-value items should be moved closer to cashier or tagged to reduce shrinkage",
			UpsideValue:   scaledCurrency(2100, seed.salesScale),
			Confidence:    0.71,
			DurationTH:    "1 สัปดาห์",
			DurationEN:    "1 week",
			TargetTH:      "ผู้จัดการและแคชเชียร์",
			TargetEN:      "Manager and cashier",
			Type:          "loss_prevention",
		},
		{
			ID: seed.suggestionPrefix + "promo4", StoreID: storeID,
			Kind: "promo", Icon: "coffee",
			TitleTH:       "โปรกาแฟกระป๋องช่วงเช้า",
			TitleEN:       "Morning canned coffee promotion",
			DescriptionTH: "กาแฟกระป๋องขายดีช่วง 07:00-09:00 แนะนำทำโปรซื้อคู่กับขนมปังหรือแซนด์วิช",
			DescriptionEN: "Canned coffee sells strongly from 07:00-09:00. Bundle with bread or sandwiches",
			UpsideValue:   scaledCurrency(2680, seed.salesScale),
			Confidence:    0.76,
			DurationTH:    "2 สัปดาห์",
			DurationEN:    "2 weeks",
			TargetTH:      "ลูกค้าก่อนเข้างาน",
			TargetEN:      "Morning commuters",
			Type:          "bundle",
		},
		{
			ID: seed.suggestionPrefix + "promo5", StoreID: storeID,
			Kind: "promo", Icon: "shopping-basket",
			TitleTH:       "เพิ่มยอดต่อบิลด้วยสินค้าหน้าเคาน์เตอร์",
			TitleEN:       "Increase basket size with counter items",
			DescriptionTH: "วางหมากฝรั่ง ลูกอม แบตเตอรี่ และทิชชู่เปียกบริเวณแคชเชียร์ เพื่อเพิ่มสินค้า impulse purchase",
			DescriptionEN: "Place gum, candy, batteries, and wet wipes near cashier to increase impulse purchases",
			UpsideValue:   scaledCurrency(3100, seed.salesScale),
			Confidence:    0.82,
			DurationTH:    "1 เดือน",
			DurationEN:    "1 month",
			TargetTH:      "ลูกค้าทุกคน",
			TargetEN:      "All customers",
			Type:          "merchandising",
		},
		{
			ID: seed.suggestionPrefix + "cust1", StoreID: storeID,
			Kind: "customer", Icon: "heart",
			TitleTH:       "เพิ่มสินค้าเพื่อสุขภาพในชั้นวางหลัก",
			TitleEN:       "Increase healthy products on main shelf",
			DescriptionTH: "กลุ่มลูกค้าในพื้นที่ซื้อโยเกิร์ต สลัด และน้ำแร่เพิ่มขึ้น แนะนำเพิ่มพื้นที่สินค้าเพื่อสุขภาพ",
			DescriptionEN: "Local customers are buying more yoghurt, salad, and mineral water. Increase healthy product shelf space",
			UpsideValue:   scaledCurrency(4200, seed.salesScale),
			Confidence:    0.73,
			DurationTH:    "1 เดือน",
			DurationEN:    "1 month",
			TargetTH:      "ลูกค้ารักสุขภาพ",
			TargetEN:      "Health-conscious customers",
			Type:          "assortment",
		},
		{
			ID: seed.suggestionPrefix + "cust2", StoreID: storeID,
			Kind: "customer", Icon: "map-pin",
			TitleTH:       "โปรเฉพาะลูกค้าในรัศมี 2 กม.",
			TitleEN:       "Local promo for customers within 2km",
			DescriptionTH: "ลูกค้าเดลิเวอรี่ส่วนใหญ่อยู่ในรัศมี 2 กม. แนะนำคูปองส่งฟรีขั้นต่ำ ฿299",
			DescriptionEN: "Most delivery customers are within 2km. Offer free delivery coupon for orders above ฿299",
			UpsideValue:   scaledCurrency(5300, seed.salesScale),
			Confidence:    0.77,
			DurationTH:    "10 วัน",
			DurationEN:    "10 days",
			TargetTH:      "ลูกค้าเดลิเวอรี่ใกล้ร้าน",
			TargetEN:      "Nearby delivery customers",
			Type:          "local_promo",
		},
	}

	if err := tx.Clauses(upsertByColumns([]string{"id"})).Create(&rows).Error; err != nil {
		return fmt.Errorf("upsert suggestions: %w", err)
	}

	return nil
}

func scaledCurrency(value float64, scale float64) float64 {
	return math.Round(value*scale*100) / 100
}

func scaledQuantity(value int, scale float64) int {
	return int(math.Round(float64(value) * scale))
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
