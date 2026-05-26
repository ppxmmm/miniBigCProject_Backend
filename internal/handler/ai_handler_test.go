package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/service"
)

type fakeRoleStoreAccessService struct {
	storeID string
	err     error
}

func (service *fakeRoleStoreAccessService) StoreIDForRole(string) (string, error) {
	return service.storeID, service.err
}

type fakeDashboardContextService struct {
	context string
	err     error
	gotID   string
}

func (service *fakeDashboardContextService) BuildContext(_ context.Context, storeID string, _ string) (string, error) {
	service.gotID = storeID
	return service.context, service.err
}

type fakeAIService struct {
	reply      string
	err        error
	gotContext string
	gotMessage string
}

func (service *fakeAIService) AskGemini(_ context.Context, question string, dashboardContext string) (string, error) {
	service.gotMessage = question
	service.gotContext = dashboardContext
	return service.reply, service.err
}

func TestAIHandlerChatValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "invalid json",
			body:       "{",
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"request body must be valid JSON\"}\n",
		},
		{
			name:       "message is required",
			body:       `{"message":" ","role":"store_manager"}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"message is required\"}\n",
		},
		{
			name:       "role is required",
			body:       `{"message":"How are sales?","role":" "}`,
			wantStatus: http.StatusBadRequest,
			wantBody:   "{\"error\":\"role is required\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := NewAIHandler(
				&fakeRoleStoreAccessService{storeID: "store_001"},
				&fakeDashboardContextService{context: "safe context"},
				&fakeAIService{reply: "reply"},
			)
			request := httptest.NewRequest(http.MethodPost, "/api/ai/chat", strings.NewReader(test.body))
			recorder := httptest.NewRecorder()

			handler.Chat(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if recorder.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.wantBody)
			}
		})
	}
}

func TestAIHandlerChatSuccess(t *testing.T) {
	t.Parallel()

	contextService := &fakeDashboardContextService{context: "sales summary only"}
	aiService := &fakeAIService{reply: "Sales are low because the weakest hour is 10:00."}
	handler := NewAIHandler(
		&fakeRoleStoreAccessService{storeID: "store_001"},
		contextService,
		aiService,
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ai/chat",
		strings.NewReader(`{"message":"Why are sales low today?","role":"store_manager"}`),
	)
	recorder := httptest.NewRecorder()

	handler.Chat(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if contextService.gotID != "store_001" {
		t.Fatalf("store id = %q, want store_001", contextService.gotID)
	}
	if aiService.gotMessage != "Why are sales low today?" {
		t.Fatalf("message = %q", aiService.gotMessage)
	}
	if aiService.gotContext != "sales summary only" {
		t.Fatalf("context = %q", aiService.gotContext)
	}

	var got map[string]string
	if err := json.Unmarshal(recorder.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["reply"] != "Sales are low because the weakest hour is 10:00." {
		t.Fatalf("reply = %q", got["reply"])
	}
}

func TestAIHandlerChatErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		accessErr  error
		contextErr error
		aiErr      error
		wantStatus int
		wantBody   string
	}{
		{
			name:       "forbidden role",
			accessErr:  service.ErrForbiddenRole,
			wantStatus: http.StatusForbidden,
			wantBody:   "{\"error\":\"role is not allowed to access this store\"}\n",
		},
		{
			name:       "store access not found",
			contextErr: service.ErrStoreAccessNotFound,
			wantStatus: http.StatusForbidden,
			wantBody:   "{\"error\":\"store access not found\"}\n",
		},
		{
			name:       "missing Gemini API key",
			aiErr:      service.ErrMissingGeminiAPIKey,
			wantStatus: http.StatusInternalServerError,
			wantBody:   "{\"error\":\"GEMINI_API_KEY is not configured\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := NewAIHandler(
				&fakeRoleStoreAccessService{storeID: "store_001", err: test.accessErr},
				&fakeDashboardContextService{context: "safe context", err: test.contextErr},
				&fakeAIService{reply: "reply", err: test.aiErr},
			)
			request := httptest.NewRequest(
				http.MethodPost,
				"/api/ai/chat",
				strings.NewReader(`{"message":"How are sales?","role":"unknown"}`),
			)
			recorder := httptest.NewRecorder()

			handler.Chat(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if recorder.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.wantBody)
			}
		})
	}
}
