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
	BuildContext(ctx context.Context, storeID string, question string) (string, error)
}

type dashboardContextService struct {
	dashboardService DashboardService
	storeIDs         map[string]int64
}

// NewDashboardContextService creates a DashboardContextService.
func NewDashboardContextService(dashboardService DashboardService) DashboardContextService {
	return &dashboardContextService{
		dashboardService: dashboardService,
		storeIDs: map[string]int64{
			"store_001": 1,
		},
	}
}

func (service *dashboardContextService) BuildContext(ctx context.Context, storeID string, _ string) (string, error) {
	dashboardStoreID, ok := service.storeIDs[storeID]
	if !ok {
		return "", ErrStoreAccessNotFound
	}

	data, err := service.dashboardService.GetDashboardByStoreID(ctx, dashboardStoreID)
	if errors.Is(err, repo.ErrNotFound) {
		return "", ErrStoreAccessNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get dashboard context for %s: %w", storeID, err)
	}

	var builder strings.Builder
	builder.WriteString("Dashboard data scope:\n")
	builder.WriteString(fmt.Sprintf("- Store access ID: %s\n", storeID))
	builder.WriteString(fmt.Sprintf("- Store: %s (%s)\n\n", data.Store.NameEN, data.Store.Code))
	writeSalesSummary(&builder, data)
	writeOrderSummary(&builder, data.Deliveries)
	writeInventorySummary(&builder, data)
	writePromotionSummary(&builder, data.Suggestions)

	return builder.String(), nil
}

func writeSalesSummary(builder *strings.Builder, data model.DashboardData) {
	builder.WriteString("Sales summary:\n")
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

func writePromotionSummary(builder *strings.Builder, suggestions []model.Suggestion) {
	builder.WriteString("Promotion and event summary:\n")
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
