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
	accessService := service.NewMockAccessService()
	dashboardContextService := service.NewDashboardContextService(dashboardService)
	var mcpClient service.MCPToolClient
	if config.AI.MCP.Enabled {
		mcpClient = service.NewStdioMCPClient(
			config.AI.MCP.Command,
			config.AI.MCP.Args,
			config.AI.MCP.BackendBaseURL,
			config.AI.MCP.Timeout,
		)
	}
	aiService := service.NewAIServiceWithMCPAndTimeout(config.AI.GeminiAPIKey, config.AI.GeminiModel, mcpClient, config.AI.Timeout)
	aiHandler := handler.NewAIHandler(accessService, dashboardContextService, aiService)
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
	r.Post("/api/ai/chat", aiHandler.Chat)
	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/ai/chat", aiHandler.Chat)

		shared := r.With(middleware.RequireRole("manager", "staff"))
		shared.Get("/dashboard", dashboardHandler.GetDefaultDashboard)
		shared.Get("/stores/{storeID}/dashboard", dashboardHandler.GetDashboardByStoreID)
		shared.Get("/stores/{storeID}", dashboardHandler.GetStoreByID)
		shared.Get("/stores/{storeID}/inventory-items", dashboardHandler.ListStoreInventoryItems)
		shared.Get("/stores/{storeID}/expiring-inventory", dashboardHandler.ListStoreExpiringInventory)
		shared.Get("/stores/{storeID}/low-stock-alerts", dashboardHandler.ListStoreLowStockAlerts)
		shared.Get("/stores/{storeID}/deliveries", dashboardHandler.ListStoreDeliveries)
		shared.Get("/stores", dataHandler.ListStores)
		shared.Get("/categories", dataHandler.ListCategories)
		shared.Get("/products", dataHandler.ListProducts)
		shared.Get("/inventory-items", dataHandler.ListInventoryItems)
		shared.Get("/expiring-inventory", dataHandler.ListExpiringInventory)
		shared.Get("/low-stock-alerts", dataHandler.ListLowStockAlerts)
		shared.Get("/deliveries", dataHandler.ListDeliveries)

		managerOnly := r.With(middleware.RequireRole("manager"))
		managerOnly.Get("/stores/{storeID}/sales/hourly", dashboardHandler.ListStoreHourlySales)
		managerOnly.Get("/stores/{storeID}/sales/daily", dashboardHandler.ListStoreDailySales)
		managerOnly.Get("/stores/{storeID}/sales/monthly", dashboardHandler.ListStoreMonthlySales)
		managerOnly.Get("/stores/{storeID}/category-sales", dashboardHandler.ListStoreCategorySales)
		managerOnly.Get("/stores/{storeID}/payment-mix", dashboardHandler.ListStorePaymentMix)
		managerOnly.Get("/stores/{storeID}/top-products", dashboardHandler.ListStoreTopProducts)
		managerOnly.Get("/stores/{storeID}/suggestions", dashboardHandler.ListStoreSuggestions)
		managerOnly.Get("/payment-methods", dataHandler.ListPaymentMethods)
		managerOnly.Get("/sales/hourly", dataHandler.ListHourlySales)
		managerOnly.Get("/sales/daily", dataHandler.ListDailySales)
		managerOnly.Get("/sales/monthly", dataHandler.ListMonthlySales)
		managerOnly.Get("/category-sales", dataHandler.ListCategorySales)
		managerOnly.Get("/payment-mix", dataHandler.ListPaymentMix)
		managerOnly.Get("/top-products", dataHandler.ListTopProducts)
		managerOnly.Get("/suggestions", dataHandler.ListSuggestions)
	})

	return r
}
