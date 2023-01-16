package models

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/crypto/bcrypt"
	"finance/internal/utils"
)

// UserID is indentifier for User
type UserID string

// NilUserID is an empty UserID
var NilUserID UserID

// User is structure represent User object
type User struct {
	ID           UserID     `json:"id,omitempty" db:"user_id"`
	Email        *string    `json:"email" db:"email"`
	PasswordHash *[]byte    `json:"-" db:"password_hash"`
	CreatedAt    *time.Time `json:"-" db:"created_at"`
	DeletedAt    *time.Time `json:"-" db:"deleted_at"`
}

// Verify all required fields before create or update
func (u *User) Verify() error {
	if u.Email == nil || (u.Email != nil && len(*u.Email) == 0) {
		return errors.New("Email is requered")
	}

	return nil
}

// Set Password updates a user's password
func (u *User) SetPassword(password string) error {
	// call hash function
	hash, err := HashPassword(password)
 	utils.CheckError(err)
	
	u.PasswordHash = &hash
	return nil
}

// CheckPassword verifies user's password
func (u *User) CheckPassword(password string) error {
	if u.PasswordHash != nil && len(*u.PasswordHash) == 0 {
		return errors.New("Password not set")
	}
	return bcrypt.CompareHashAndPassword(*u.PasswordHash, []byte(password))
}

// HashPassword hashes a user's raw password
func HashPassword(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
}
