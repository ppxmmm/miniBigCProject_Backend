package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"google.golang.org/genai"
)

func TestAIServiceMissingGeminiAPIKey(t *testing.T) {
	t.Parallel()

	service := NewAIService("", "")
	_, err := service.AskGemini(context.Background(), "question", "context", RoleAccess{})
	if !errors.Is(err, ErrMissingGeminiAPIKey) {
		t.Fatalf("error = %v, want %v", err, ErrMissingGeminiAPIKey)
	}
}

func TestAIServiceOperationContextIgnoresParentCancellation(t *testing.T) {
	t.Parallel()

	parent, cancelParent := context.WithCancel(context.Background())
	cancelParent()

	service := &geminiAIService{timeout: time.Minute}
	ctx, cancel := service.operationContext(parent)
	defer cancel()

	select {
	case <-ctx.Done():
		t.Fatalf("operation context should not inherit parent cancellation: %v", ctx.Err())
	default:
	}
	if _, ok := ctx.Deadline(); !ok {
		t.Fatal("operation context missing deadline")
	}
}

func TestWrapGeminiErrorClassifiesContextErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "deadline", err: context.DeadlineExceeded, want: ErrAIRequestTimeout},
		{name: "canceled", err: context.Canceled, want: ErrAIRequestCanceled},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := wrapGeminiError("generate Gemini content", test.err)
			if !errors.Is(err, test.want) {
				t.Fatalf("error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestIsTransientGeminiError(t *testing.T) {
	t.Parallel()

	err := genai.APIError{Code: 503, Status: "UNAVAILABLE", Message: "high demand"}
	if !isTransientGeminiError(err) {
		t.Fatal("503 UNAVAILABLE should be transient")
	}
	if isTransientGeminiError(errors.New("invalid api key")) {
		t.Fatal("auth errors should not be transient")
	}
}

func TestIsQuotaExceededError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "api status resource exhausted",
			err:  genai.APIError{Code: 429, Status: "RESOURCE_EXHAUSTED", Message: "quota exceeded"},
			want: true,
		},
		{
			name: "wrapped api status resource exhausted",
			err:  fmt.Errorf("generate content: %w", genai.APIError{Code: 429, Status: "RESOURCE_EXHAUSTED"}),
			want: true,
		},
		{
			name: "message mentions tokens",
			err:  errors.New("input token limit exceeded"),
			want: true,
		},
		{
			name: "ordinary service error",
			err:  errors.New("temporary upstream failure"),
			want: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := isQuotaExceededError(test.err); got != test.want {
				t.Fatalf("isQuotaExceededError() = %v, want %v", got, test.want)
			}
		})
	}
}

func TestModelContentWithFunctionCallsPreservesThoughtSignature(t *testing.T) {
	t.Parallel()

	signature := []byte("signed-thought")
	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Role: string(genai.RoleModel),
					Parts: []*genai.Part{
						{
							Thought:          true,
							ThoughtSignature: signature,
							Text:             "private reasoning",
						},
						{
							FunctionCall: &genai.FunctionCall{
								ID:   "call-1",
								Name: "get_store_data",
								Args: map[string]any{"store_id": float64(1), "endpoint": "suggestions"},
							},
							ThoughtSignature: signature,
						},
					},
				},
			},
		},
	}

	content := modelContentWithFunctionCalls(response)
	if content == nil {
		t.Fatal("modelContentWithFunctionCalls() returned nil")
	}
	if content.Role != string(genai.RoleModel) {
		t.Fatalf("role = %q, want %q", content.Role, genai.RoleModel)
	}
	if len(content.Parts) != 2 {
		t.Fatalf("parts length = %d, want 2", len(content.Parts))
	}
	if string(content.Parts[0].ThoughtSignature) != string(signature) {
		t.Fatalf("thought signature was not preserved")
	}
	if string(content.Parts[1].ThoughtSignature) != string(signature) {
		t.Fatalf("function call thought signature was not preserved")
	}
}

func TestNewFunctionResponsePartPreservesFunctionCallID(t *testing.T) {
	t.Parallel()

	part := newFunctionResponsePart(
		&genai.FunctionCall{ID: "call-1", Name: "get_store_data"},
		map[string]any{"ok": true},
	)

	if part.FunctionResponse == nil {
		t.Fatal("FunctionResponse is nil")
	}
	if part.FunctionResponse.ID != "call-1" {
		t.Fatalf("id = %q, want call-1", part.FunctionResponse.ID)
	}
	if part.FunctionResponse.Name != "get_store_data" {
		t.Fatalf("name = %q, want get_store_data", part.FunctionResponse.Name)
	}
	if part.FunctionResponse.Response["ok"] != true {
		t.Fatalf("response = %#v, want ok=true", part.FunctionResponse.Response)
	}
}

func TestCollectResponseTextSkipsThoughtAndFunctionParts(t *testing.T) {
	t.Parallel()

	response := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Thought: true, Text: "private reasoning"},
						{Text: "public answer"},
					},
				},
			},
		},
	}

	got := collectResponseText(response)
	if got != "public answer" {
		t.Fatalf("collectResponseText() = %q, want public answer", got)
	}
}

func TestSystemInstructionForMode(t *testing.T) {
	t.Parallel()

	if strings.Contains(systemInstructionForMode(false), "Call the most relevant available MCP tool") {
		t.Fatal("prefetch-only instruction should not ask Gemini to call tools")
	}
	for _, want := range []string{
		"Call the most relevant available MCP tool",
		"must call at least one MCP tool",
	} {
		if !strings.Contains(systemInstructionForMode(true), want) {
			t.Fatalf("tool-calling instruction missing %q", want)
		}
	}
}

func TestExtractGeminiReply(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		response  *genai.GenerateContentResponse
		wantReply string
		wantErr   error
	}{
		{
			name: "text reply",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{{Text: "  answer  "}},
						},
					},
				},
			},
			wantReply: "answer",
		},
		{
			name: "unresolved function call",
			response: &genai.GenerateContentResponse{
				Candidates: []*genai.Candidate{
					{
						Content: &genai.Content{
							Parts: []*genai.Part{
								{
									FunctionCall: &genai.FunctionCall{
										Name: "get_store_data",
										Args: map[string]any{"store_id": float64(1)},
									},
								},
							},
						},
					},
				},
			},
			wantErr: ErrUnresolvedGeminiFunctionCalls,
		},
		{
			name:     "empty response",
			response: &genai.GenerateContentResponse{},
			wantErr:  ErrEmptyAIResponse,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := extractGeminiReply(test.response)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if got != test.wantReply {
				t.Fatalf("reply = %q, want %q", got, test.wantReply)
			}
		})
	}
}

func TestAccessAllowsMCPTool(t *testing.T) {
	t.Parallel()

	staff := RoleAccess{Role: "staff", StoreAccessID: "store_001", DashboardStoreID: 1}
	manager := RoleAccess{Role: "manager", StoreAccessID: "store_001", DashboardStoreID: 1, CanViewManagerData: true}

	if accessAllowsMCPTool(staff, "get_store_dashboard") {
		t.Fatal("staff should not be allowed to use manager dashboard tool")
	}
	if accessAllowsMCPTool(staff, "list_suggestions") {
		t.Fatal("staff should not be allowed to use suggestions tool")
	}
	if !accessAllowsMCPTool(staff, "get_store_data") {
		t.Fatal("staff should be allowed to use scoped store data tool")
	}
	if !accessAllowsMCPTool(manager, "get_store_dashboard") {
		t.Fatal("manager should be allowed to use store dashboard tool")
	}
	if accessAllowsMCPTool(manager, "get_global_data") {
		t.Fatal("global data tools should not be exposed through scoped AI")
	}
}

func TestAuthorizeMCPToolCall(t *testing.T) {
	t.Parallel()

	staff := RoleAccess{Role: "staff", StoreAccessID: "store_001", DashboardStoreID: 1}

	authorized, err := authorizeMCPToolCall(staff, "get_store_data", map[string]any{
		"store_id": float64(1),
		"endpoint": "low-stock-alerts",
	})
	if err != nil {
		t.Fatalf("authorize allowed call: %v", err)
	}
	if authorized["store_id"] != int64(1) {
		t.Fatalf("store_id = %#v, want int64(1)", authorized["store_id"])
	}

	_, err = authorizeMCPToolCall(staff, "get_store_data", map[string]any{
		"store_id": float64(1),
		"endpoint": "sales/daily",
	})
	if err == nil {
		t.Fatal("staff sales endpoint should be forbidden")
	}

	_, err = authorizeMCPToolCall(staff, "get_store", map[string]any{"store_id": float64(2)})
	if err == nil {
		t.Fatal("other store should be forbidden")
	}

	authorized, err = authorizeMCPToolCall(staff, "list_deliveries", map[string]any{})
	if err != nil {
		t.Fatalf("authorize optional scoped call: %v", err)
	}
	if authorized["store_id"] != int64(1) {
		t.Fatalf("store_id = %#v, want int64(1)", authorized["store_id"])
	}
}

func TestRestrictMCPInputSchema(t *testing.T) {
	t.Parallel()

	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"endpoint": map[string]any{
				"type": "string",
				"enum": []any{"sales/daily", "low-stock-alerts", "suggestions"},
			},
		},
	}

	restrictMCPInputSchema(RoleAccess{Role: "staff"}, "get_store_data", schema)

	properties := schema["properties"].(map[string]any)
	endpoint := properties["endpoint"].(map[string]any)
	got := endpoint["enum"].([]any)
	if len(got) != 4 {
		t.Fatalf("staff endpoint enum length = %d, want 4: %#v", len(got), got)
	}
	for _, restricted := range []string{"sales/daily", "suggestions"} {
		for _, value := range got {
			if value == restricted {
				t.Fatalf("staff endpoint enum contains restricted value %q: %#v", restricted, got)
			}
		}
	}
}

func TestDescribeMCPToolAddsDashboardDataHints(t *testing.T) {
	t.Parallel()

	description := describeMCPTool(MCPTool{
		Name:        "get_store_dashboard",
		Description: "Return dashboard data.",
	})

	for _, want := range []string{"broad", "diagnosis", "multiple dashboard areas"} {
		if !strings.Contains(description, want) {
			t.Fatalf("description missing %q: %s", want, description)
		}
	}
}

func TestSystemInstructionRequiresMCPForDataQuestions(t *testing.T) {
	t.Parallel()

	for _, want := range []string{
		"must call at least one MCP tool",
		"website data catalog",
		"Call the most relevant available MCP tool",
		"authorized numeric MCP store_id",
		"broad, ambiguous, causal",
		"Never answer that a website metric does not exist",
		"secretary brief",
		"operations secretary",
	} {
		if !strings.Contains(systemInstruction, want) {
			t.Fatalf("systemInstruction missing %q", want)
		}
	}
}
