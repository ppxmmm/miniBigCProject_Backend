package router

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/handler"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/middleware"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/service"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/util"
	"gorm.io/gorm"
)

// New creates the HTTP router for the backend.
func New(appName string, database *gorm.DB, config util.Config) http.Handler {
	dashboardRepository := repo.NewDashboardRepository(database)
	dashboardService := service.NewDashboardService(dashboardRepository)
	dashboardHandler := handler.NewDashboardHandler(dashboardService)
	dataRepository := repo.NewDataRepository(database)
	dataService := service.NewDataService(dataRepository)
	dataHandler := handler.NewDataHandler(dataService)

	r := chi.NewRouter()
	r.Use(middleware.CORS(config.CORSOrigins))
	r.Get("/", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "Welcome to %s", appName)
	})
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		sqlDB, err := database.DB()
		if err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}
		if err := sqlDB.PingContext(r.Context()); err != nil {
			http.Error(w, "database unavailable", http.StatusServiceUnavailable)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "ok")
	})
	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/dashboard", dashboardHandler.GetDefaultDashboard)
		r.Get("/stores/{storeID}/dashboard", dashboardHandler.GetDashboardByStoreID)
		r.Get("/stores/{storeID}", dashboardHandler.GetStoreByID)
		r.Get("/stores/{storeID}/sales/hourly", dashboardHandler.ListStoreHourlySales)
		r.Get("/stores/{storeID}/sales/daily", dashboardHandler.ListStoreDailySales)
		r.Get("/stores/{storeID}/sales/monthly", dashboardHandler.ListStoreMonthlySales)
		r.Get("/stores/{storeID}/category-sales", dashboardHandler.ListStoreCategorySales)
		r.Get("/stores/{storeID}/payment-mix", dashboardHandler.ListStorePaymentMix)
		r.Get("/stores/{storeID}/top-products", dashboardHandler.ListStoreTopProducts)
		r.Get("/stores/{storeID}/inventory-items", dashboardHandler.ListStoreInventoryItems)
		r.Get("/stores/{storeID}/expiring-inventory", dashboardHandler.ListStoreExpiringInventory)
		r.Get("/stores/{storeID}/low-stock-alerts", dashboardHandler.ListStoreLowStockAlerts)
		r.Get("/stores/{storeID}/deliveries", dashboardHandler.ListStoreDeliveries)
		r.Get("/stores/{storeID}/suggestions", dashboardHandler.ListStoreSuggestions)
		r.Get("/stores", dataHandler.ListStores)
		r.Get("/categories", dataHandler.ListCategories)
		r.Get("/payment-methods", dataHandler.ListPaymentMethods)
		r.Get("/products", dataHandler.ListProducts)
		r.Get("/sales/hourly", dataHandler.ListHourlySales)
		r.Get("/sales/daily", dataHandler.ListDailySales)
		r.Get("/sales/monthly", dataHandler.ListMonthlySales)
		r.Get("/category-sales", dataHandler.ListCategorySales)
		r.Get("/payment-mix", dataHandler.ListPaymentMix)
		r.Get("/top-products", dataHandler.ListTopProducts)
		r.Get("/inventory-items", dataHandler.ListInventoryItems)
		r.Get("/expiring-inventory", dataHandler.ListExpiringInventory)
		r.Get("/low-stock-alerts", dataHandler.ListLowStockAlerts)
		r.Get("/deliveries", dataHandler.ListDeliveries)
		r.Get("/suggestions", dataHandler.ListSuggestions)
	})

	return r
}
