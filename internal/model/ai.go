package model

// AIChatRequest is the request body for the AI chat endpoint.
type AIChatRequest struct {
	Message string `json:"message"`
	Role    string `json:"role"`
}

// AIChatResponse is the response body for the AI chat endpoint.
type AIChatResponse struct {
	Reply string `json:"reply"`
}
