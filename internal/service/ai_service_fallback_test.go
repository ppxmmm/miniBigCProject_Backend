package service

import (
	"strings"
	"testing"
)

func TestBuildFallbackReplySales(t *testing.T) {
	t.Parallel()

	context := `Dashboard data scope:
- Store access ID: store_001
- Store: Mini BigC · Thonglor 13 (MBC-0421)

Sales summary:
- Latest daily sales: THB 296000.00, comparison THB 274000.00, change THB 22000.00 (8.0%)
- Weakest hour: 05:00 with THB 1200.00 sales; strongest hour: 19:00 with THB 21400.00 sales
- Top categories: Food & Beverage THB 142800.00 share 38.0% trend 6.2%

Order summary:
- Delivery orders: 20, total order value THB 12345.00, late orders 3
`

	reply := buildFallbackReply("Why are sales low today?", context)
	if strings.Contains(strings.ToLower(reply), "unavailable") {
		t.Fatalf("fallback reply should not mention unavailable: %s", reply)
	}
	if !strings.Contains(reply, "Latest daily sales") {
		t.Fatalf("fallback reply should mention sales data: %s", reply)
	}
}

func TestBuildFallbackReplyInventoryThai(t *testing.T) {
	t.Parallel()

	context := `Dashboard data scope:
- Store access ID: store_001

Inventory and low-stock summary:
- Inventory items visible: 13
- Low-stock alerts: Milk stock 2 reorder 20; Bread stock 4 reorder 18
- Expiring inventory: Bread expires 2026-05-27 quantity 10
`

	reply := buildFallbackReply("สต็อกอะไรน่ากังวลที่สุด", context)
	if !strings.Contains(reply, "สรุปสต็อก") {
		t.Fatalf("fallback reply should be in Thai and mention inventory: %s", reply)
	}
	if !strings.Contains(reply, "low-stock") {
		t.Fatalf("fallback reply should include low-stock details: %s", reply)
	}
}
