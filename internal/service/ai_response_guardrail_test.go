package service

import (
	"errors"
	"testing"
)

func TestValidateAIResponseGuardrails(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		access   RoleAccess
		decision QuestionAuthorizationDecision
		reply    string
		wantErr  error
	}{
		{
			name:     "staff allowed inventory answer",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{Domains: []string{"low_stock"}},
			reply:    "Milk is running low. Please restock shelf A first.",
		},
		{
			name:     "staff sales leakage blocked",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{Domains: []string{"low_stock"}},
			reply:    "Sales are THB 12,000 today.",
			wantErr:  ErrUnsafeAIResponse,
		},
		{
			name:     "staff thai sales leakage blocked",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{Domains: []string{"operational_actions"}},
			reply:    "ยอดขายวันนี้ 12,000 บาท ต่ำกว่าเป้า",
			wantErr:  ErrUnsafeAIResponse,
		},
		{
			name:     "staff safe disclaimer allowed",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{Domains: []string{"operational_actions"}},
			reply:    "วันนี้ให้โฟกัสเติมของใกล้หมดและเช็กออเดอร์ล่าช้า ข้อมูลยอดขายและเป้าเป็นข้อมูลผู้จัดการค่ะ",
		},
		{
			name:     "staff operational answer with order value allowed",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{Domains: []string{"deliveries"}},
			reply:    "มีออเดอร์ล่าช้า 2 รายการ รวมมูลค่า THB 500 ให้ติดตามคนขับก่อนค่ะ",
		},
		{
			name:     "staff restricted decision blocked",
			access:   RoleAccess{Role: "staff"},
			decision: QuestionAuthorizationDecision{RequiresManagerData: true, Domains: []string{"sales_performance"}},
			reply:    "I cannot answer that.",
			wantErr:  ErrUnsafeAIResponse,
		},
		{
			name:     "manager bypasses output guardrail",
			access:   RoleAccess{Role: "manager", CanViewManagerData: true},
			decision: QuestionAuthorizationDecision{RequiresManagerData: true, Domains: []string{"sales_performance"}},
			reply:    "Sales are THB 12,000 today.",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := validateAIResponseGuardrails(test.access, test.decision, test.reply)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error = %v, want %v", err, test.wantErr)
			}
		})
	}
}
