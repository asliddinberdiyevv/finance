package auth

import "finance/internal/models"

// we will have 3 permission  type for now.
type PermissionTypes string

const (
	// User is loged in (we have userID in principal)
	Admin PermissionTypes = "admin"
	// User is loged in (we have userID in principal)
	Member PermissionTypes = "member"
	// User is loged in and user id passed to API is the same
	MemberIsTarget PermissionTypes = "member_is_target"
)

// We will create functions for each type

// Admin
var adminOnly = func(roles []*models.UserRole) bool {
	for _, role := range roles {
		switch role.Role {
		case models.RoleAdmin:
			return true
		}
	}
	return false
}

// Loged in user
var member = func(principal models.Principal) bool {
	return principal.UserID != ""
}

// Loged in user = Target user
var memberIsTarget = func(userID models.UserID, principal models.Principal) bool {
	if userID == "" || principal.UserID == "" {
		return false
	}
	if userID != principal.UserID {
		return false
	}
	return true
}
