package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
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
	storeID, err := strconv.ParseInt(chi.URLParam(r, "storeID"), 10, 64)
	if err != nil || storeID <= 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "storeID must be a positive integer"})
		return
	}

	data, err := handler.service.GetDashboardByStoreID(r.Context(), storeID)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, data)
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
