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
	access service.RoleAccess
	err    error
}

func (service *fakeRoleStoreAccessService) AccessForRole(string) (service.RoleAccess, error) {
	return service.access, service.err
}

type fakeDashboardContextService struct {
	context string
	err     error
	gotRole string
	gotID   string
}

func (service *fakeDashboardContextService) BuildContext(_ context.Context, access service.RoleAccess, _ string) (string, error) {
	service.gotRole = access.Role
	service.gotID = access.StoreAccessID
	return service.context, service.err
}

type fakeAIService struct {
	reply      string
	err        error
	gotContext string
	gotMessage string
}

func (service *fakeAIService) AskGemini(_ context.Context, question string, dashboardContext string, _ service.RoleAccess) (string, error) {
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
				&fakeRoleStoreAccessService{access: service.RoleAccess{Role: "manager", StoreAccessID: "store_001"}},
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
		&fakeRoleStoreAccessService{access: service.RoleAccess{Role: "store_manager", StoreAccessID: "store_001", DashboardStoreID: 1, CanViewManagerData: true}},
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
	if contextService.gotRole != "store_manager" {
		t.Fatalf("role = %q, want store_manager", contextService.gotRole)
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

func TestAIHandlerChatUsesRoleHeaderFallback(t *testing.T) {
	t.Parallel()

	contextService := &fakeDashboardContextService{context: "sales summary only"}
	handler := NewAIHandler(
		&fakeRoleStoreAccessService{access: service.RoleAccess{Role: "manager", StoreAccessID: "store_001", DashboardStoreID: 1, CanViewManagerData: true}},
		contextService,
		&fakeAIService{reply: "reply"},
	)
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/ai/chat",
		strings.NewReader(`{"message":"How are sales?"}`),
	)
	request.Header.Set("X-Frontend-Role", "manager")
	recorder := httptest.NewRecorder()

	handler.Chat(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if contextService.gotRole != "manager" {
		t.Fatalf("role = %q, want manager", contextService.gotRole)
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
		{
			name:       "empty AI response",
			aiErr:      service.ErrEmptyAIResponse,
			wantStatus: http.StatusBadGateway,
			wantBody:   "{\"error\":\"AI response is empty\"}\n",
		},
		{
			name:       "unresolved Gemini function calls",
			aiErr:      service.ErrUnresolvedGeminiFunctionCalls,
			wantStatus: http.StatusBadGateway,
			wantBody:   "{\"error\":\"Gemini function calls were not resolved\"}\n",
		},
		{
			name:       "AI request timeout",
			aiErr:      service.ErrAIRequestTimeout,
			wantStatus: http.StatusGatewayTimeout,
			wantBody:   "{\"error\":\"AI request timed out\"}\n",
		},
		{
			name:       "AI request canceled",
			aiErr:      service.ErrAIRequestCanceled,
			wantStatus: http.StatusRequestTimeout,
			wantBody:   "{\"error\":\"AI request canceled\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			handler := NewAIHandler(
				&fakeRoleStoreAccessService{
					access: service.RoleAccess{Role: "manager", StoreAccessID: "store_001", DashboardStoreID: 1, CanViewManagerData: true},
					err:    test.accessErr,
				},
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
