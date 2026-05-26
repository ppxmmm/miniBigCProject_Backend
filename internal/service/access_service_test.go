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
		wantError error
	}{
		{name: "store manager", role: "store_manager", wantID: "store_001"},
		{name: "manager", role: "manager", wantID: "store_001"},
		{name: "staff", role: "staff", wantID: "store_001"},
		{name: "admin", role: "admin", wantID: "store_001"},
		{name: "trims and normalizes", role: " Store_Manager ", wantID: "store_001"},
		{name: "unknown role forbidden", role: "cashier", wantError: ErrForbiddenRole},
	}

	service := NewMockAccessService()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, err := service.StoreIDForRole(test.role)
			if !errors.Is(err, test.wantError) {
				t.Fatalf("error = %v, want %v", err, test.wantError)
			}
			if got != test.wantID {
				t.Fatalf("store id = %q, want %q", got, test.wantID)
			}
		})
	}
}
