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
	if request.Role == "" {
		request.Role = roleFromAIRequestHeader(r)
	}
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

	request.History = sanitizeAIChatRequestHistory(request.History)
	contextualMessage := request.Message
	if len(request.History) > 0 {
		contextualMessage = request.Message + "\n\nRecent conversation:\n" + formatAIChatRequestHistory(request.History)
	}

	dashboardContext, err := handler.dashboardContextService.BuildContext(r.Context(), access, contextualMessage)
	if err != nil {
		writeAIError(w, err)
		return
	}

	reply, err := handler.aiService.AskGemini(r.Context(), request.Message, dashboardContext, access, request.History)
	if err != nil {
		writeAIError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, model.AIChatResponse{Reply: reply})
}

func sanitizeAIChatRequestHistory(history []model.AIChatMessage) []model.AIChatMessage {
	if len(history) == 0 {
		return nil
	}

	const maxHistoryMessages = 10
	if len(history) > maxHistoryMessages {
		history = history[len(history)-maxHistoryMessages:]
	}

	sanitized := make([]model.AIChatMessage, 0, len(history))
	for _, message := range history {
		role := strings.ToLower(strings.TrimSpace(message.Role))
		if role != "user" && role != "assistant" {
			continue
		}
		content := strings.TrimSpace(message.Content)
		if content == "" {
			continue
		}
		const maxHistoryContentLength = 1200
		if len(content) > maxHistoryContentLength {
			content = content[:maxHistoryContentLength] + "... truncated"
		}
		sanitized = append(sanitized, model.AIChatMessage{Role: role, Content: content})
	}

	return sanitized
}

func formatAIChatRequestHistory(history []model.AIChatMessage) string {
	var builder strings.Builder
	for i, message := range history {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString("- ")
		builder.WriteString(message.Role)
		builder.WriteString(": ")
		builder.WriteString(message.Content)
	}
	return builder.String()
}

func roleFromAIRequestHeader(r *http.Request) string {
	for _, header := range []string{"X-User-Role", "X-Frontend-Role", "X-Role"} {
		if role := strings.TrimSpace(r.Header.Get(header)); role != "" {
			return role
		}
	}
	return ""
}

func writeAIError(w http.ResponseWriter, err error) {
	log.Printf("AI request failed: %v", err)

	switch {
	case errors.Is(err, service.ErrForbiddenRole):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "role is not allowed to access this store"})
	case errors.Is(err, service.ErrUnauthorizedManagerData):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "manager-only data is not authorized for this role"})
	case errors.Is(err, service.ErrUnclearQuestionAuthorization):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "question authorization is unclear"})
	case errors.Is(err, service.ErrUnsafeAIResponse):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "AI response did not pass authorization guardrails"})
	case errors.Is(err, service.ErrStoreAccessNotFound):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "store access not found"})
	case errors.Is(err, service.ErrMissingGeminiAPIKey):
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "GEMINI_API_KEY is not configured"})
	case errors.Is(err, service.ErrEmptyAIResponse):
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "AI response is empty"})
	case errors.Is(err, service.ErrUnresolvedGeminiFunctionCalls):
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Gemini function calls were not resolved"})
	case errors.Is(err, service.ErrAIRequestTimeout):
		writeJSON(w, http.StatusGatewayTimeout, map[string]string{"error": "AI request timed out"})
	case errors.Is(err, service.ErrAIRequestCanceled):
		writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "AI request canceled"})
	default:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "AI service unavailable"})
	}
}
