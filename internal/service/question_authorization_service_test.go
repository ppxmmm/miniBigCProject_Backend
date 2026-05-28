package service

import (
	"context"
	"errors"
	"testing"
)

type fallbackQuestionAuthorizer struct {
	decision QuestionAuthorizationDecision
	err      error
	called   bool
}

func (authorizer *fallbackQuestionAuthorizer) ClassifyQuestion(context.Context, string, RoleAccess) (QuestionAuthorizationDecision, error) {
	authorizer.called = true
	return authorizer.decision, authorizer.err
}

func TestDeterministicQuestionAuthorizationService(t *testing.T) {
	t.Parallel()

	service := NewDeterministicQuestionAuthorizationService()
	tests := []struct {
		name        string
		question    string
		wantManager bool
		wantDomain  string
		wantErr     error
	}{
		{
			name:        "money made is sales performance",
			question:    "How much money did we make today?",
			wantManager: true,
			wantDomain:  "sales_performance",
		},
		{
			name:       "restock is low stock",
			question:   "What products should we restock first?",
			wantDomain: "low_stock",
		},
		{
			name:       "late delivery is delivery",
			question:   "Any late deliveries today?",
			wantDomain: "deliveries",
		},
		{
			name:       "thai action items are operational actions",
			question:   "รายการที่ต้องดำเนินการมีอะไรบ้าง",
			wantDomain: "operational_actions",
		},
		{
			name:       "thai broad improvement is operational actions",
			question:   "วันนี้มีอะไรที่ต้องปรับปรุง",
			wantDomain: "operational_actions",
		},
		{
			name:       "thai should improve question is operational actions",
			question:   "ควรปรับปรุงอะไรมั้ย",
			wantDomain: "operational_actions",
		},
		{
			name:       "thai product availability is staff allowed",
			question:   "มีนมมั้ย",
			wantDomain: "product_availability",
		},
		{
			name:        "thai sales improvement remains manager only",
			question:    "ยอดขายวันนี้มีอะไรที่ต้องปรับปรุง",
			wantManager: true,
			wantDomain:  "sales_performance",
		},
		{
			name:        "thai revenue availability tie breaks manager only",
			question:    "มีรายได้ไหม",
			wantManager: true,
			wantDomain:  "sales_performance",
		},
		{
			name:     "unclear question falls through",
			question: "Can you check this?",
			wantErr:  ErrUnclearQuestionAuthorization,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.ClassifyQuestion(t.Context(), test.question, RoleAccess{Role: "staff"})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}
			if got.RequiresManagerData != test.wantManager {
				t.Fatalf("RequiresManagerData = %v, want %v", got.RequiresManagerData, test.wantManager)
			}
			if len(got.Domains) == 0 || got.Domains[0] != test.wantDomain {
				t.Fatalf("domains = %#v, want first domain %q", got.Domains, test.wantDomain)
			}
			if got.Source != "semantic_deterministic" {
				t.Fatalf("source = %q, want semantic_deterministic", got.Source)
			}
		})
	}
}

func TestCompositeQuestionAuthorizationServiceFallback(t *testing.T) {
	t.Parallel()

	fallback := &fallbackQuestionAuthorizer{
		decision: QuestionAuthorizationDecision{
			RequiresManagerData: true,
			Domains:             []string{"sales_performance"},
			Confidence:          0.88,
			Source:              "fake",
		},
	}
	service := &compositeQuestionAuthorizationService{
		deterministic: NewDeterministicQuestionAuthorizationService(),
		fallback:      fallback,
	}

	decision, err := service.ClassifyQuestion(t.Context(), "Is the store healthy?", RoleAccess{Role: "staff"})
	if err != nil {
		t.Fatalf("ClassifyQuestion returned error: %v", err)
	}
	if !fallback.called {
		t.Fatal("fallback classifier was not called for unclear deterministic decision")
	}
	if !decision.RequiresManagerData {
		t.Fatal("fallback decision should require manager data")
	}
}

func TestCompositeQuestionAuthorizationServiceMissingFallbackKeyIsUnclear(t *testing.T) {
	t.Parallel()

	service := &compositeQuestionAuthorizationService{
		deterministic: NewDeterministicQuestionAuthorizationService(),
		fallback:      NewGeminiQuestionAuthorizationService("", "", defaultAITimeout),
	}

	_, err := service.ClassifyQuestion(t.Context(), "Is the store healthy?", RoleAccess{Role: "staff"})
	if !errors.Is(err, ErrUnclearQuestionAuthorization) {
		t.Fatalf("error = %v, want %v", err, ErrUnclearQuestionAuthorization)
	}
}

func TestParseQuestionAuthorizationDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		input       string
		wantManager bool
		wantDomain  string
		wantErr     error
	}{
		{
			name: "plain json",
			input: `{
				"requires_manager_data": true,
				"domains": ["sales_performance"],
				"confidence": 0.94,
				"reason": "The question asks about money made today."
			}`,
			wantManager: true,
			wantDomain:  "sales_performance",
		},
		{
			name: "markdown json",
			input: "```json\n" + `{
				"requires_manager_data": false,
				"domains": ["low_stock"],
				"confidence": 0.91,
				"reason": "The question asks about replenishment."
			}` + "\n```",
			wantDomain: "low_stock",
		},
		{
			name:    "missing json",
			input:   "staff can ask this",
			wantErr: ErrUnclearQuestionAuthorization,
		},
		{
			name:    "invalid json",
			input:   `{"requires_manager_data":`,
			wantErr: ErrUnclearQuestionAuthorization,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := parseQuestionAuthorizationDecision(test.input)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
			if test.wantErr != nil {
				return
			}
			if got.RequiresManagerData != test.wantManager {
				t.Fatalf("RequiresManagerData = %v, want %v", got.RequiresManagerData, test.wantManager)
			}
			if len(got.Domains) == 0 || got.Domains[0] != test.wantDomain {
				t.Fatalf("domains = %#v, want first domain %q", got.Domains, test.wantDomain)
			}
		})
	}
}

func TestGeminiQuestionAuthorizationServiceAllowsManagersWithoutAPIKey(t *testing.T) {
	t.Parallel()

	service := NewGeminiQuestionAuthorizationService("", "", defaultAITimeout)
	decision, err := service.ClassifyQuestion(t.Context(), "How much money did we make today?", RoleAccess{
		Role:               "manager",
		CanViewManagerData: true,
	})
	if err != nil {
		t.Fatalf("ClassifyQuestion returned error: %v", err)
	}
	if decision.RequiresManagerData {
		t.Fatal("manager-access decision should not require a manager-data denial")
	}
}

func TestGeminiQuestionAuthorizationServiceRequiresAPIKeyForStaff(t *testing.T) {
	t.Parallel()

	service := NewGeminiQuestionAuthorizationService("", "", defaultAITimeout)
	_, err := service.ClassifyQuestion(t.Context(), "How much money did we make today?", RoleAccess{Role: "staff"})
	if !errors.Is(err, ErrMissingGeminiAPIKey) {
		t.Fatalf("error = %v, want %v", err, ErrMissingGeminiAPIKey)
	}
}
