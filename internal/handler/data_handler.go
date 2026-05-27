package handler

import (
	"context"
	"net/http"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/service"
)

// DataHandler handles read-only raw data API routes.
type DataHandler struct {
	service service.DataService
}

// NewDataHandler creates a DataHandler.
func NewDataHandler(service service.DataService) *DataHandler {
	return &DataHandler{service: service}
}

// ListStores returns all store records.
func (handler *DataHandler) ListStores(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListStores)
}

// ListCategories returns all category records.
func (handler *DataHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListCategories)
}

// ListPaymentMethods returns all payment method records.
func (handler *DataHandler) ListPaymentMethods(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListPaymentMethods)
}

// ListProducts returns all product records.
func (handler *DataHandler) ListProducts(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListProducts)
}

// ListHourlySales returns all hourly sales records.
func (handler *DataHandler) ListHourlySales(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListHourlySales)
}

// ListDailySales returns all daily sales records.
func (handler *DataHandler) ListDailySales(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListDailySales)
}

// ListMonthlySales returns all monthly sales records.
func (handler *DataHandler) ListMonthlySales(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListMonthlySales)
}

// ListCategorySales returns all category sales records.
func (handler *DataHandler) ListCategorySales(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListCategorySales)
}

// ListPaymentMix returns all payment mix records.
func (handler *DataHandler) ListPaymentMix(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListPaymentMix)
}

// ListTopProducts returns all top product records.
func (handler *DataHandler) ListTopProducts(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListTopProducts)
}

// ListInventoryItems returns all inventory item records.
func (handler *DataHandler) ListInventoryItems(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListInventoryItems)
}

// ListExpiringInventory returns all expiring inventory records.
func (handler *DataHandler) ListExpiringInventory(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListExpiringInventory)
}

// ListLowStockAlerts returns all low stock alert records.
func (handler *DataHandler) ListLowStockAlerts(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListLowStockAlerts)
}

// ListDeliveries returns all delivery records.
func (handler *DataHandler) ListDeliveries(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListDeliveries)
}

// ListSuggestions returns all suggestion records.
func (handler *DataHandler) ListSuggestions(w http.ResponseWriter, r *http.Request) {
	writeList(w, r, handler.service.ListSuggestions)
}

func writeList[T any](w http.ResponseWriter, r *http.Request, loader func(context.Context) ([]T, error)) {
	rows, err := loader(r.Context())
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, rows)
}
