package util

import (
	"strings"
	"testing"
	"time"
)

func TestLoadConfigDefaultCORSOriginsIncludesViteDevServer(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "")

	config := LoadConfig()

	for _, want := range []string{"http://localhost:5173", "http://127.0.0.1:5173"} {
		if !strings.Contains(config.CORSOrigins, want) {
			t.Fatalf("default CORS origins missing %q: %s", want, config.CORSOrigins)
		}
	}
}

func TestLoadConfigAITimeout(t *testing.T) {
	t.Setenv("AI_TIMEOUT_SECONDS", "120")

	config := LoadConfig()

	if config.AI.Timeout != 120*time.Second {
		t.Fatalf("AI timeout = %s, want 120s", config.AI.Timeout)
	}
}
