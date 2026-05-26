package service

import (
	"context"
	"errors"
	"testing"
)

func TestAIServiceMissingGeminiAPIKey(t *testing.T) {
	t.Parallel()

	service := NewAIService("", "")
	_, err := service.AskGemini(context.Background(), "question", "context")
	if !errors.Is(err, ErrMissingGeminiAPIKey) {
		t.Fatalf("error = %v, want %v", err, ErrMissingGeminiAPIKey)
	}
}
