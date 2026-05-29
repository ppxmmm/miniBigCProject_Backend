package model

// AIChatRequest is the request body for the AI chat endpoint.
type AIChatRequest struct {
	Message string          `json:"message"`
	Role    string          `json:"role"`
	History []AIChatMessage `json:"history,omitempty"`
}

// AIChatMessage is one prior chat turn sent by the client for follow-up context.
type AIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIChatResponse is the response body for the AI chat endpoint.
type AIChatResponse struct {
	Reply string `json:"reply"`
}
