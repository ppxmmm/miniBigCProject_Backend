package service

import (
	"errors"
	"testing"
)

func TestMockAccessServiceStoreIDForRole(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		role      string
		wantID    string
		wantRole  string
		wantMgr   bool
		wantError error
	}{
		{name: "store manager", role: "store_manager", wantID: "store_001", wantRole: "store_manager", wantMgr: true},
		{name: "manager", role: "manager", wantID: "store_001", wantRole: "manager", wantMgr: true},
		{name: "staff", role: "staff", wantID: "store_001", wantRole: "staff"},
		{name: "admin", role: "admin", wantID: "store_001", wantRole: "admin", wantMgr: true},
		{name: "trims and normalizes", role: " Store_Manager ", wantID: "store_001", wantRole: "store_manager", wantMgr: true},
		{name: "unknown role forbidden", role: "cashier", wantError: ErrForbiddenRole},
	}

	service := NewMockAccessService()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.AccessForRole(test.role)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
			if got.StoreAccessID != test.wantID {
				t.Fatalf("store id = %q, want %q", got.StoreAccessID, test.wantID)
			}
			if got.Role != test.wantRole {
				t.Fatalf("role = %q, want %q", got.Role, test.wantRole)
			}
			if got.CanViewManagerData != test.wantMgr {
				t.Fatalf("CanViewManagerData = %v, want %v", got.CanViewManagerData, test.wantMgr)
			}
		})
	}
}
