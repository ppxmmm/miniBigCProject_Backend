package service

import (
	"errors"
	"strings"
)

// ErrForbiddenRole is returned when a prototype role has no store access.
var ErrForbiddenRole = errors.New("role is not allowed to access a store")

// RoleStoreAccessService maps a frontend mock role to a backend-controlled store ID.
type RoleStoreAccessService interface {
	AccessForRole(role string) (RoleAccess, error)
}

type mockAccessService struct{}

// RoleAccess is the backend-controlled data scope for a prototype role.
type RoleAccess struct {
	Role                 string
	StoreAccessID        string
	DashboardStoreID     int64
	CanViewManagerData   bool
	AllowedDataSummaries []string
}

// NewMockAccessService creates prototype role-to-store access control.
func NewMockAccessService() RoleStoreAccessService {
	return &mockAccessService{}
}

func (service *mockAccessService) AccessForRole(role string) (RoleAccess, error) {
	// Temporary prototype access control. Later replace this with real backend
	// auth/JWT/session and user-store permissions from the database.
	normalizedRole := strings.ToLower(strings.TrimSpace(role))
	switch normalizedRole {
	case "store_manager", "manager", "admin":
		return RoleAccess{
			Role:               normalizedRole,
			StoreAccessID:      "store_001",
			DashboardStoreID:   1,
			CanViewManagerData: true,
			AllowedDataSummaries: []string{
				"sales",
				"orders",
				"inventory",
				"promotions",
			},
		}, nil
	case "staff":
		return RoleAccess{
			Role:               normalizedRole,
			StoreAccessID:      "store_001",
			DashboardStoreID:   1,
			CanViewManagerData: false,
			AllowedDataSummaries: []string{
				"sales",
				"orders",
				"inventory",
				"promotions",
			},
		}, nil
	default:
		return RoleAccess{}, ErrForbiddenRole
	}
}
