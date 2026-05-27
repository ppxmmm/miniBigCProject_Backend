package handler

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/ppxmmm/miniBigCProject_Backend/internal/model"
	"github.com/ppxmmm/miniBigCProject_Backend/internal/service"
)

// AIHandler handles AI chat routes.
type AIHandler struct {
	accessService           service.RoleStoreAccessService
	dashboardContextService service.DashboardContextService
	aiService               service.AIService
}

// NewAIHandler creates an AIHandler.
func NewAIHandler(
	accessService service.RoleStoreAccessService,
	dashboardContextService service.DashboardContextService,
	aiService service.AIService,
) *AIHandler {
	return &AIHandler{
		accessService:           accessService,
		dashboardContextService: dashboardContextService,
		aiService:               aiService,
	}
}

// Chat answers dashboard questions using Gemini and store-scoped context.
func (handler *AIHandler) Chat(w http.ResponseWriter, r *http.Request) {
	var request model.AIChatRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "request body must be valid JSON"})
		return
	}

	request.Message = strings.TrimSpace(request.Message)
	request.Role = strings.TrimSpace(request.Role)
	if request.Message == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "message is required"})
		return
	}
	if request.Role == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "role is required"})
		return
	}

	access, err := handler.accessService.AccessForRole(request.Role)
	if err != nil {
		writeAIError(w, err)
		return
	}

	dashboardContext, err := handler.dashboardContextService.BuildContext(r.Context(), access, request.Message)
	if err != nil {
		writeAIError(w, err)
		return
	}

	reply, err := handler.aiService.AskGemini(r.Context(), request.Message, dashboardContext, access)
	if err != nil {
		writeAIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, model.AIChatResponse{Reply: reply})
}

func writeAIError(w http.ResponseWriter, err error) {
	log.Printf("AI request failed: %v", err)

	switch {
	case errors.Is(err, service.ErrForbiddenRole):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "role is not allowed to access this store"})
	case errors.Is(err, service.ErrStoreAccessNotFound):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "store access not found"})
	case errors.Is(err, service.ErrMissingGeminiAPIKey):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "GEMINI_API_KEY is not configured"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "AI service unavailable"})
	}
}
