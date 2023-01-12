package database

import (
	"context"
	"finance/internal/models"
)

type SessionsDB interface {
	SaveRefreshToken(ctx context.Context, session models.Session) error
}

var insertOrUpdateSession = `
	INSERT INTO sessions (user_id, device_id, refresh_token, expires_at)
	VALUES(:user_id, :device_id, :refresh_token, :expires_at)

	ON CONFLICT (user_id, device_id)
	DO 
		UPDATE
			SET refresh_token = :refresh_token,
					expires_at = :expires_at;

`

func (d *database) SaveRefreshToken(ctx context.Context, session models.Session) error {
	if _, err := d.conn.NamedQueryContext(ctx, insertOrUpdateSession, session); err != nil {
		return err
	}
	return nil
}
