package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestAIServiceMissingGeminiAPIKey(t *testing.T) {
	t.Parallel()

	service := NewAIService("", "")
	_, err := service.AskGemini(context.Background(), "question", "context", RoleAccess{})
	if !errors.Is(err, ErrMissingGeminiAPIKey) {
		t.Fatalf("error = %v, want %v", err, ErrMissingGeminiAPIKey)
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
		"decide for yourself what data is needed",
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

func TestBuildDashboardFallbackReplyImprovementQuestion(t *testing.T) {
	t.Parallel()

	context := `Dashboard data scope:
- Role: manager

Sales summary:
- MTD sales: THB 900.00, dashboard target/comparison line THB 1200.00, gap THB -300.00 (-25.0%). In the website, MTD target means the backend daily comparison series.
- Top categories: Food THB 700.00 share 70.0% trend -4.0%

Order summary:
- Delivery orders: 4, total order value THB 1000.00, late orders 1

Inventory and low-stock summary:
- Inventory items visible: 10
- Low-stock alerts: Milk stock 2 reorder 20

Promotion and event summary:
- promo: Markdown bread; upside THB 250.00; confidence 80%; duration 1 day`

	reply := buildDashboardFallbackReply("ควรปรับปรุงอะไรสำหรับเดือนนี้", context)

	for _, want := range []string{
		"สรุปคือเดือนนี้ควรโฟกัส 3 เรื่องค่ะ",
		"กู้ยอดขายจากช่องว่าง MTD",
		"THB -300.00",
		"Milk stock 2 reorder 20",
		"Markdown bread",
	} {
		if !strings.Contains(reply, want) {
			t.Fatalf("fallback reply missing %q:\n%s", want, reply)
		}
	}
}
