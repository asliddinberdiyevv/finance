package models

// Credentials used in login API
type Credentials struct {
	// Username/Password login:
	Email    string `json:"email"`
	Password string `json:"password"`
}
