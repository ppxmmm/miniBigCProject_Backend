package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/repo"
)

type fakeDashboardService struct {
	data       model.DashboardData
	err        error
	gotStoreID int64
}

func (service *fakeDashboardService) GetDefaultDashboard(context.Context) (model.DashboardData, error) {
	return service.data, service.err
}

func (service *fakeDashboardService) GetDashboardByStoreID(_ context.Context, storeID int64) (model.DashboardData, error) {
	service.gotStoreID = storeID
	return service.data, service.err
}

func TestDashboardHandlerStoreScopedRoutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		handler        func(*DashboardHandler, http.ResponseWriter, *http.Request)
		data           model.DashboardData
		wantStatus     int
		wantStoreID    int64
		wantBody       string
		wantBodyAssert func(*testing.T, []byte)
	}{
		{
			name:        "store by id returns store",
			handler:     (*DashboardHandler).GetStoreByID,
			data:        model.DashboardData{Store: model.Store{ID: 7, Code: "BKK-007"}},
			wantStatus:  http.StatusOK,
			wantStoreID: 7,
			wantBodyAssert: func(t *testing.T, body []byte) {
				t.Helper()
				var got model.Store
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if got.ID != 7 || got.Code != "BKK-007" {
					t.Fatalf("unexpected store response: %+v", got)
				}
			},
		},
		{
			name:    "store top products returns selected dashboard slice",
			handler: (*DashboardHandler).ListStoreTopProducts,
			data: model.DashboardData{
				TopProducts: []model.TopProduct{{ID: 10, StoreID: 7, SKU: "SKU-1", SalesValue: 99}},
			},
			wantStatus:  http.StatusOK,
			wantStoreID: 7,
			wantBodyAssert: func(t *testing.T, body []byte) {
				t.Helper()
				var got []model.TopProduct
				if err := json.Unmarshal(body, &got); err != nil {
					t.Fatalf("decode response: %v", err)
				}
				if len(got) != 1 || got[0].SKU != "SKU-1" {
					t.Fatalf("unexpected top products response: %+v", got)
				}
			},
		},
		{
			name:        "invalid store id returns bad request",
			handler:     (*DashboardHandler).ListStoreSuggestions,
			wantStatus:  http.StatusBadRequest,
			wantStoreID: 0,
			wantBody:    "{\"error\":\"storeID must be a positive integer\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &fakeDashboardService{data: test.data}
			handler := NewDashboardHandler(service)
			request := httptest.NewRequest(http.MethodGet, "/stores/7", nil)
			storeID := "7"
			if test.wantStatus == http.StatusBadRequest {
				storeID = "abc"
			}
			request = request.WithContext(withURLParam(request.Context(), "storeID", storeID))
			recorder := httptest.NewRecorder()

			test.handler(handler, recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if service.gotStoreID != test.wantStoreID {
				t.Fatalf("store id = %d, want %d", service.gotStoreID, test.wantStoreID)
			}
			if test.wantBody != "" && recorder.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.wantBody)
			}
			if test.wantBodyAssert != nil {
				test.wantBodyAssert(t, recorder.Body.Bytes())
			}
		})
	}
}

func TestDashboardHandlerStoreScopedRouteErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "not found maps to 404",
			err:        repo.ErrNotFound,
			wantStatus: http.StatusNotFound,
			wantBody:   "{\"error\":\"record not found\"}\n",
		},
		{
			name:       "unexpected error maps to 500",
			err:        errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError,
			wantBody:   "{\"error\":\"internal server error\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			service := &fakeDashboardService{err: test.err}
			handler := NewDashboardHandler(service)
			request := httptest.NewRequest(http.MethodGet, "/stores/7/deliveries", nil)
			request = request.WithContext(withURLParam(request.Context(), "storeID", "7"))
			recorder := httptest.NewRecorder()

			handler.ListStoreDeliveries(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if recorder.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.wantBody)
			}
		})
	}
}

func withURLParam(ctx context.Context, key string, value string) context.Context {
	routeContext := chi.NewRouteContext()
	routeContext.URLParams.Add(key, value)
	return context.WithValue(ctx, chi.RouteCtxKey, routeContext)
}
