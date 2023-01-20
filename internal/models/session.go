package models

import "github.com/pkg/errors"

type DeviceID string

var NilDeviceID DeviceID

// Session is represent user's sessions
type Session struct {
	UserID       UserID   `db:"user_id"`
	DeviceID     DeviceID `db:"device_id"`
	RefreshToken string   `db:"refresh_token"`
	ExpiresAt    int64    `db:"expires_at"`
}

// SessionData used to represent data sent in json body with requests
type SessionData struct {
	DeviceID DeviceID `json:"device_id"`
}

// Verify all required fields before create or update
func (u *SessionData) Verify() error {
	if len(u.DeviceID) == 0 {
		return errors.New("DeviceID is requered")
	}

	return nil
}
