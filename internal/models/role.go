package models

// Role is a function a user can serve
type Role string

const (
	// RoleAdmin is a an administrator of App. Root
	RoleAdmin Role = "admin"
)

type UserRole struct {
	Role Role `json:"role" db:"role"`
}
