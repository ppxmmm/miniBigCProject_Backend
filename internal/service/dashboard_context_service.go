package service

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
)

// ErrStoreAccessNotFound is returned when a prototype store ID is not allowed.
var ErrStoreAccessNotFound = errors.New("store access not found")

// DashboardContextService builds a safe store-scoped prompt context.
type DashboardContextService interface {
	BuildContext(ctx context.Context, access RoleAccess, question string) (string, error)
}

type dashboardContextService struct {
	dashboardService DashboardService
}

// NewDashboardContextService creates a DashboardContextService.
func NewDashboardContextService(dashboardService DashboardService) DashboardContextService {
	return &dashboardContextService{
		dashboardService: dashboardService,
	}
}

func (service *dashboardContextService) BuildContext(ctx context.Context, access RoleAccess, _ string) (string, error) {
	if access.StoreAccessID == "" || access.DashboardStoreID == 0 {
		return "", ErrStoreAccessNotFound
	}

	data, err := service.dashboardService.GetDashboardByStoreID(ctx, access.DashboardStoreID)
	if errors.Is(err, repo.ErrNotFound) {
		return "", ErrStoreAccessNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get dashboard context for %s: %w", access.StoreAccessID, err)
	}

	var builder strings.Builder
	builder.WriteString("Dashboard data scope:\n")
	builder.WriteString(fmt.Sprintf("- Role: %s\n", access.Role))
	builder.WriteString(fmt.Sprintf("- Store access ID: %s\n", access.StoreAccessID))
	builder.WriteString(fmt.Sprintf("- Authorized MCP store_id: %d\n", access.DashboardStoreID))
	builder.WriteString(fmt.Sprintf("- Store: %s (%s)\n\n", data.Store.NameEN, data.Store.Code))
	builder.WriteString("Role permissions:\n")
	builder.WriteString(fmt.Sprintf("- Allowed data summaries: %s\n", strings.Join(access.AllowedDataSummaries, ", ")))
	if !access.CanViewManagerData {
		builder.WriteString("- MCP tool access is limited for this role. Revenue, sales, payment mix, top products, category sales, promotion, and suggestion data are not authorized here.\n")
	}
	builder.WriteString("\n")

	writeWebsiteDataCatalog(&builder, access)
	writeSalesSummary(&builder, access, data)
	writeOrderSummary(&builder, data.Deliveries)
	writeInventorySummary(&builder, data)
	writePromotionSummary(&builder, access, data.Suggestions)

	return builder.String(), nil
}

func writeWebsiteDataCatalog(builder *strings.Builder, access RoleAccess) {
	builder.WriteString("Website data catalog for MCP tool selection:\n")
	if access.CanViewManagerData {
		builder.WriteString("- dashboard: broad store snapshot with sales, categories, payment mix, top products, inventory, deliveries, and suggestions. Use for open-ended questions or cross-functional diagnosis.\n")
		builder.WriteString("- sales/hourly: intraday sales and comparison by hour. Use for today, traffic, weak hours, peak hours, staffing, and daypart questions.\n")
		builder.WriteString("- sales/daily: daily sales and comparison series. Use for MTD, target/comparison line, pacing, recovery gap, week/month trend, and questions about why sales are below target.\n")
		builder.WriteString("- sales/monthly: monthly sales series. Use for YTD, latest month, monthly average, and longer trend questions.\n")
		builder.WriteString("- category-sales: category revenue, share, and trend. Use for which categories are driving or dragging sales.\n")
		builder.WriteString("- payment-mix: payment method share. Use for cash, QR, card, wallet, and checkout behavior questions.\n")
		builder.WriteString("- top-products: top SKU sales, quantity, and trend. Use for item gain/loss, hero SKUs, product movers, and basket opportunities.\n")
		builder.WriteString("- suggestions: backend actions, promotions, events, risk, inventory, and customer opportunities. Use for recommended next actions and recovery plans.\n")
	} else {
		builder.WriteString("- dashboard, sales/hourly, sales/daily, sales/monthly, category-sales, payment-mix, top-products, suggestions: manager-only data and unavailable to this role.\n")
	}
	builder.WriteString("- inventory-items: product stock position. Use for stock quantity, location, price, and store inventory checks.\n")
	builder.WriteString("- low-stock-alerts: SKUs below reorder level. Use for OOS risk and replenishment priorities.\n")
	builder.WriteString("- expiring-inventory: near-expiry stock. Use for aging, shrink, markdown, and waste questions.\n")
	builder.WriteString("- deliveries: order workload and fulfillment status. Use for delivery, open orders, late orders, OTIF-style operational questions.\n")
	builder.WriteString("\n")
}

func writeSalesSummary(builder *strings.Builder, access RoleAccess, data model.DashboardData) {
	builder.WriteString("Sales summary:\n")
	if !access.CanViewManagerData {
		builder.WriteString("- Sales, revenue, category, payment mix, and top-product metrics are restricted for this role.\n\n")
		return
	}
	if len(data.DailySales) > 0 {
		var mtdSales float64
		var mtdTarget float64
		for _, sale := range data.DailySales {
			mtdSales += sale.SalesValue
			mtdTarget += sale.ComparisonSalesValue
		}
		gap := mtdSales - mtdTarget
		percent := percentChange(mtdSales, mtdTarget)
		builder.WriteString(fmt.Sprintf(
			"- MTD sales: THB %.2f, dashboard target/comparison line THB %.2f, gap THB %.2f (%.1f%%). In the website, MTD target means the backend daily comparison series.\n",
			mtdSales,
			mtdTarget,
			gap,
			percent,
		))
	}
	if len(data.DailySales) > 0 {
		latest := data.DailySales[len(data.DailySales)-1]
		delta := latest.SalesValue - latest.ComparisonSalesValue
		percent := percentChange(latest.SalesValue, latest.ComparisonSalesValue)
		builder.WriteString(fmt.Sprintf(
			"- Latest daily sales: THB %.2f, comparison THB %.2f, change THB %.2f (%.1f%%)\n",
			latest.SalesValue,
			latest.ComparisonSalesValue,
			delta,
			percent,
		))
	}
	if len(data.HourlySales) > 0 {
		weakest := data.HourlySales[0]
		strongest := data.HourlySales[0]
		for _, sale := range data.HourlySales[1:] {
			if sale.SalesValue < weakest.SalesValue {
				weakest = sale
			}
			if sale.SalesValue > strongest.SalesValue {
				strongest = sale
			}
		}
		builder.WriteString(fmt.Sprintf(
			"- Weakest hour: %02d:00 with THB %.2f sales; strongest hour: %02d:00 with THB %.2f sales\n",
			weakest.Hour,
			weakest.SalesValue,
			strongest.Hour,
			strongest.SalesValue,
		))
	}
	if len(data.CategorySales) > 0 {
		categories := append([]model.CategorySale(nil), data.CategorySales...)
		sort.Slice(categories, func(i int, j int) bool {
			return categories[i].SalesValue > categories[j].SalesValue
		})
		limit := min(3, len(categories))
		builder.WriteString("- Top categories: ")
		for i := range limit {
			if i > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(fmt.Sprintf(
				"%s THB %.2f share %.1f%% trend %.1f%%",
				categories[i].NameEN,
				categories[i].SalesValue,
				categories[i].Share*100,
				categories[i].TrendPercent,
			))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func writeOrderSummary(builder *strings.Builder, deliveries []model.Delivery) {
	builder.WriteString("Order summary:\n")
	statusCounts := make(map[string]int)
	var totalOrderValue float64
	lateCount := 0
	for _, delivery := range deliveries {
		statusCounts[delivery.Status]++
		totalOrderValue += delivery.OrderValue
		if delivery.IsLate {
			lateCount++
		}
	}
	builder.WriteString(fmt.Sprintf("- Delivery orders: %d, total order value THB %.2f, late orders %d\n", len(deliveries), totalOrderValue, lateCount))
	if len(statusCounts) > 0 {
		statuses := make([]string, 0, len(statusCounts))
		for status := range statusCounts {
			statuses = append(statuses, status)
		}
		sort.Strings(statuses)
		builder.WriteString("- Status counts: ")
		for i, status := range statuses {
			if i > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(fmt.Sprintf("%s %d", status, statusCounts[status]))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func writeInventorySummary(builder *strings.Builder, data model.DashboardData) {
	builder.WriteString("Inventory and low-stock summary:\n")
	builder.WriteString(fmt.Sprintf("- Inventory items visible: %d\n", len(data.InventoryItems)))
	if len(data.LowStockAlerts) > 0 {
		limit := min(5, len(data.LowStockAlerts))
		builder.WriteString("- Low-stock alerts: ")
		for i := range limit {
			alert := data.LowStockAlerts[i]
			if i > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(fmt.Sprintf("%s stock %d reorder %d", alert.NameEN, alert.StockQuantity, alert.ReorderQuantity))
		}
		builder.WriteString("\n")
	}
	if len(data.ExpiringInventory) > 0 {
		limit := min(5, len(data.ExpiringInventory))
		builder.WriteString("- Expiring inventory: ")
		for i := range limit {
			item := data.ExpiringInventory[i]
			if i > 0 {
				builder.WriteString("; ")
			}
			builder.WriteString(fmt.Sprintf("%s expires %s quantity %d", item.NameEN, item.ExpiryDate.Format("2006-01-02"), item.StockQuantity))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func writePromotionSummary(builder *strings.Builder, access RoleAccess, suggestions []model.Suggestion) {
	builder.WriteString("Promotion and event summary:\n")
	if !access.CanViewManagerData {
		builder.WriteString("- Promotion, suggestion, and business recovery recommendations are restricted for this role.\n")
		return
	}
	if len(suggestions) == 0 {
		builder.WriteString("- No promotion or event suggestions are available.\n")
		return
	}
	limit := min(5, len(suggestions))
	for i := range limit {
		suggestion := suggestions[i]
		builder.WriteString(fmt.Sprintf(
			"- %s: %s; upside THB %.2f; confidence %.0f%%; duration %s\n",
			suggestion.Kind,
			suggestion.TitleEN,
			suggestion.UpsideValue,
			suggestion.Confidence*100,
			suggestion.DurationEN,
		))
	}
}

func percentChange(current float64, previous float64) float64 {
	if previous == 0 {
		return 0
	}
	return math.Round(((current-previous)/previous)*1000) / 10
}
