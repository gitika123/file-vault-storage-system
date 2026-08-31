package policy

import "errors"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type Principal struct {
	UserID string
	Role   Role
}

var (
	ErrUnauthenticated = errors.New("authentication required")
	ErrForbidden       = errors.New("permission denied")
)

func RequireAuthenticated(principal *Principal) error {
	if principal == nil || principal.UserID == "" {
		return ErrUnauthenticated
	}
	return nil
}

func RequireAdmin(principal *Principal) error {
	if err := RequireAuthenticated(principal); err != nil {
		return err
	}
	if principal.Role != RoleAdmin {
		return ErrForbidden
	}
	return nil
}

func CanDeleteFile(principal *Principal, uploaderID string) error {
	if err := RequireAuthenticated(principal); err != nil {
		return err
	}
	if principal.UserID != uploaderID {
		return ErrForbidden
	}
	return nil
}
