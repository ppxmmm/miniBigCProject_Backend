package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/service"
)

// DashboardHandler handles read-only dashboard API routes.
type DashboardHandler struct {
	service service.DashboardService
}

// NewDashboardHandler creates a DashboardHandler.
func NewDashboardHandler(service service.DashboardService) *DashboardHandler {
	return &DashboardHandler{service: service}
}

// GetDefaultDashboard returns all dashboard mock data for the default store.
func (handler *DashboardHandler) GetDefaultDashboard(w http.ResponseWriter, r *http.Request) {
	data, err := handler.service.GetDefaultDashboard(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// GetDashboardByStoreID returns all dashboard mock data for one store.
func (handler *DashboardHandler) GetDashboardByStoreID(w http.ResponseWriter, r *http.Request) {
	storeID, ok := parseStoreID(w, r)
	if !ok {
		return
	}

	data, err := handler.service.GetDashboardByStoreID(r.Context(), storeID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
}

// GetStoreByID returns one store record.
func (handler *DashboardHandler) GetStoreByID(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) model.Store {
		return data.Store
	})
}

// ListStoreHourlySales returns hourly sales for one store.
func (handler *DashboardHandler) ListStoreHourlySales(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.HourlySale {
		return data.HourlySales
	})
}

// ListStoreDailySales returns daily sales for one store.
func (handler *DashboardHandler) ListStoreDailySales(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.DailySale {
		return data.DailySales
	})
}

// ListStoreMonthlySales returns monthly sales for one store.
func (handler *DashboardHandler) ListStoreMonthlySales(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.MonthlySale {
		return data.MonthlySales
	})
}

// ListStoreCategorySales returns category sales for one store.
func (handler *DashboardHandler) ListStoreCategorySales(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.CategorySale {
		return data.CategorySales
	})
}

// ListStorePaymentMix returns payment mix records for one store.
func (handler *DashboardHandler) ListStorePaymentMix(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.PaymentMix {
		return data.PaymentMix
	})
}

// ListStoreTopProducts returns top product records for one store.
func (handler *DashboardHandler) ListStoreTopProducts(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.TopProduct {
		return data.TopProducts
	})
}

// ListStoreInventoryItems returns inventory item records for one store.
func (handler *DashboardHandler) ListStoreInventoryItems(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.InventoryItem {
		return data.InventoryItems
	})
}

// ListStoreExpiringInventory returns expiring inventory records for one store.
func (handler *DashboardHandler) ListStoreExpiringInventory(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.ExpiringInventoryItem {
		return data.ExpiringInventory
	})
}

// ListStoreLowStockAlerts returns low stock alert records for one store.
func (handler *DashboardHandler) ListStoreLowStockAlerts(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.LowStockAlert {
		return data.LowStockAlerts
	})
}

// ListStoreDeliveries returns delivery records for one store.
func (handler *DashboardHandler) ListStoreDeliveries(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.Delivery {
		return data.Deliveries
	})
}

// ListStoreSuggestions returns suggestion records for one store.
func (handler *DashboardHandler) ListStoreSuggestions(w http.ResponseWriter, r *http.Request) {
	writeStoreDashboardValue(w, r, handler.service, func(data model.DashboardData) []model.Suggestion {
		return data.Suggestions
	})
}

func parseStoreID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	storeID, err := strconv.ParseInt(chi.URLParam(r, "storeID"), 10, 64)
	if err != nil || storeID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "storeID must be a positive integer"})
		return 0, false
	}

	return storeID, true
}

func writeStoreDashboardValue[T any](
	w http.ResponseWriter,
	r *http.Request,
	service service.DashboardService,
	selector func(model.DashboardData) T,
) {
	storeID, ok := parseStoreID(w, r)
	if !ok {
		return
	}

	data, err := service.GetDashboardByStoreID(r.Context(), storeID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, selector(data))
}

func writeServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, repo.ErrNotFound) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "record not found"})
		return
	}

	writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}
