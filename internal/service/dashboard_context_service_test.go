package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
)

type fakeContextDashboardService struct {
	data       model.DashboardData
	err        error
	gotStoreID int64
}

func (service *fakeContextDashboardService) GetDefaultDashboard(context.Context) (model.DashboardData, error) {
	return service.data, service.err
}

func (service *fakeContextDashboardService) GetDashboardByStoreID(_ context.Context, storeID int64) (model.DashboardData, error) {
	service.gotStoreID = storeID
	return service.data, service.err
}

func TestDashboardContextServiceBuildContext(t *testing.T) {
	t.Parallel()

	expiryDate := time.Date(2026, 5, 27, 0, 0, 0, 0, time.UTC)
	dashboardService := &fakeContextDashboardService{
		data: model.DashboardData{
			Store: model.Store{NameEN: "Mini BigC Demo", Code: "MBC-0421"},
			DailySales: []model.DailySale{
				{SalesValue: 1000, ComparisonSalesValue: 1200},
			},
			HourlySales: []model.HourlySale{
				{Hour: 9, SalesValue: 100},
				{Hour: 10, SalesValue: 80},
				{Hour: 11, SalesValue: 140},
			},
			CategorySales: []model.CategorySale{
				{NameEN: "Food", SalesValue: 700, Share: 0.7, TrendPercent: -4},
			},
			InventoryItems: []model.InventoryItem{{NameEN: "Milk"}},
			LowStockAlerts: []model.LowStockAlert{
				{NameEN: "Milk", StockQuantity: 2, ReorderQuantity: 20},
			},
			ExpiringInventory: []model.ExpiringInventoryItem{
				{NameEN: "Bread", ExpiryDate: expiryDate, StockQuantity: 10},
			},
			Deliveries: []model.Delivery{
				{CustomerNameEN: "Private Customer", AddressEN: "Private Address", Status: "preparing", OrderValue: 100, IsLate: true},
			},
			Suggestions: []model.Suggestion{
				{Kind: "promo", TitleEN: "Markdown bread", UpsideValue: 250, Confidence: 0.8, DurationEN: "1 day"},
			},
		},
	}
	service := NewDashboardContextService(dashboardService)

	got, err := service.BuildContext(context.Background(), RoleAccess{
		Role:               "manager",
		StoreAccessID:      "store_001",
		DashboardStoreID:   1,
		CanViewManagerData: true,
		AllowedDataSummaries: []string{
			"sales",
			"orders",
			"inventory",
			"promotions",
		},
	}, "question")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}
	if dashboardService.gotStoreID != 1 {
		t.Fatalf("store id = %d, want 1", dashboardService.gotStoreID)
	}

	required := []string{"Sales summary:", "Order summary:", "Inventory and low-stock summary:", "Promotion and event summary:"}
	for _, text := range required {
		if !strings.Contains(got, text) {
			t.Fatalf("context missing %q:\n%s", text, got)
		}
	}
	for _, text := range []string{
		"Website data catalog for MCP tool selection:",
		"sales/daily: daily sales and comparison series",
		"top-products: top SKU sales",
		"deliveries: order workload",
	} {
		if !strings.Contains(got, text) {
			t.Fatalf("context missing data catalog detail %q:\n%s", text, got)
		}
	}
	if !strings.Contains(got, "Authorized MCP store_id: 1") {
		t.Fatalf("context missing authorized MCP store id:\n%s", got)
	}
	for _, text := range []string{
		"MTD sales: THB 1000.00",
		"dashboard target/comparison line THB 1200.00",
		"In the website, MTD target means the backend daily comparison series.",
	} {
		if !strings.Contains(got, text) {
			t.Fatalf("context missing MTD target detail %q:\n%s", text, got)
		}
	}
	forbidden := []string{"Private Customer", "Private Address"}
	for _, text := range forbidden {
		if strings.Contains(got, text) {
			t.Fatalf("context contains sensitive text %q:\n%s", text, got)
		}
	}
}

func TestDashboardContextServiceBuildContextStaffRedactsManagerMetrics(t *testing.T) {
	t.Parallel()

	dashboardService := &fakeContextDashboardService{
		data: model.DashboardData{
			Store: model.Store{NameEN: "Mini BigC Demo", Code: "MBC-0421"},
			DailySales: []model.DailySale{
				{SalesValue: 1000, ComparisonSalesValue: 1200},
			},
			CategorySales: []model.CategorySale{
				{NameEN: "Food", SalesValue: 700, Share: 0.7, TrendPercent: -4},
			},
			InventoryItems: []model.InventoryItem{{NameEN: "Milk"}},
			LowStockAlerts: []model.LowStockAlert{{NameEN: "Milk", StockQuantity: 2, ReorderQuantity: 20}},
			Deliveries:     []model.Delivery{{Status: "preparing", OrderValue: 100, IsLate: true}},
			Suggestions:    []model.Suggestion{{Kind: "promo", TitleEN: "Markdown bread", UpsideValue: 250}},
		},
	}
	service := NewDashboardContextService(dashboardService)

	got, err := service.BuildContext(context.Background(), RoleAccess{
		Role:                 "staff",
		StoreAccessID:        "store_001",
		DashboardStoreID:     1,
		CanViewManagerData:   false,
		AllowedDataSummaries: []string{"sales", "orders", "inventory", "promotions"},
	}, "question")
	if err != nil {
		t.Fatalf("BuildContext returned error: %v", err)
	}

	required := []string{
		"Role: staff",
		"MCP tool access is limited for this role.",
		"Sales summary:",
		"Sales, revenue, category, payment mix, and top-product metrics are restricted for this role.",
		"Order summary:",
		"Inventory and low-stock summary:",
		"Promotion and event summary:",
		"Promotion, suggestion, and business recovery recommendations are restricted for this role.",
	}
	for _, text := range required {
		if !strings.Contains(got, text) {
			t.Fatalf("staff context missing %q:\n%s", text, got)
		}
	}
	for _, forbidden := range []string{"Latest daily sales", "dashboard target/comparison line", "Top categories:", "Sales: THB", "Markdown bread"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("staff context leaked %q:\n%s", forbidden, got)
		}
	}
}

func TestDashboardContextServiceBuildContextErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		storeID   string
		sourceErr error
		wantErr   error
	}{
		{name: "unknown prototype store", storeID: "store_999", wantErr: ErrStoreAccessNotFound},
		{name: "repository not found maps to store access not found", storeID: "store_001", sourceErr: repo.ErrNotFound, wantErr: ErrStoreAccessNotFound},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := NewDashboardContextService(&fakeContextDashboardService{err: test.sourceErr})
			access := RoleAccess{
				Role:             "manager",
				StoreAccessID:    test.storeID,
				DashboardStoreID: 1,
			}
			if test.storeID == "store_999" {
				access.DashboardStoreID = 0
			}
			_, err := service.BuildContext(context.Background(), access, "question")
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
