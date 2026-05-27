package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		header     string
		role       string
		allowed    []string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "allows matching role",
			header:     roleHeader,
			role:       " manager ",
			allowed:    []string{"manager"},
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "allows alternate frontend header",
			header:     alternateRoleHead,
			role:       "staff",
			allowed:    []string{"manager", "staff"},
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "rejects missing role",
			allowed:    []string{"manager"},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "{\"error\":\"role is required\"}\n",
		},
		{
			name:       "rejects disallowed role",
			header:     roleHeader,
			role:       "staff",
			allowed:    []string{"manager"},
			wantStatus: http.StatusForbidden,
			wantBody:   "{\"error\":\"role is not allowed to access this resource\"}\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte("ok")); err != nil {
					t.Fatalf("write response: %v", err)
				}
			})
			handler := RequireRole(test.allowed...)(next)
			request := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
			if test.header != "" {
				request.Header.Set(test.header, test.role)
			}
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, request)

			if recorder.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", recorder.Code, test.wantStatus)
			}
			if recorder.Body.String() != test.wantBody {
				t.Fatalf("body = %q, want %q", recorder.Body.String(), test.wantBody)
			}
		})
	}
}
