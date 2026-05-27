package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
)

const (
	roleHeader        = "X-User-Role"
	alternateRoleHead = "X-Frontend-Role"
	legacyRoleHeader  = "X-Role"
)

// RequireRole allows requests from one of the provided frontend roles.
func RequireRole(allowedRoles ...string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedRoles))
	for _, role := range allowedRoles {
		normalized := normalizeRole(role)
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			role := roleFromRequest(r)
			if role == "" {
				writeMiddlewareJSON(w, http.StatusUnauthorized, map[string]string{"error": "role is required"})
				return
			}

			if _, ok := allowed[role]; !ok {
				writeMiddlewareJSON(w, http.StatusForbidden, map[string]string{"error": "role is not allowed to access this resource"})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func roleFromRequest(r *http.Request) string {
	for _, header := range []string{roleHeader, alternateRoleHead, legacyRoleHeader} {
		if role := normalizeRole(r.Header.Get(header)); role != "" {
			return role
		}
	}
	return ""
}

func normalizeRole(role string) string {
	return strings.ToLower(strings.TrimSpace(role))
}

func writeMiddlewareJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "encode response", http.StatusInternalServerError)
	}
}
