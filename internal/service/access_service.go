package service

import (
	"errors"
	"strings"
)

// ErrForbiddenRole is returned when a prototype role has no store access.
var ErrForbiddenRole = errors.New("role is not allowed to access a store")

// RoleStoreAccessService maps a frontend mock role to a backend-controlled store ID.
type RoleStoreAccessService interface {
	StoreIDForRole(role string) (string, error)
}

type mockAccessService struct{}

// NewMockAccessService creates prototype role-to-store access control.
func NewMockAccessService() RoleStoreAccessService {
	return &mockAccessService{}
}

func (service *mockAccessService) StoreIDForRole(role string) (string, error) {
	// Temporary prototype access control. Later replace this with real backend
	// auth/JWT/session and user-store permissions from the database.
	switch strings.ToLower(strings.TrimSpace(role)) {
	case "store_manager", "manager", "staff", "admin":
		return "store_001", nil
	default:
		return "", ErrForbiddenRole
	}
}
