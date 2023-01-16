package database

import (
	"context"
	"finance/internal/models"
	"finance/internal/utils"
)

type SessionsDB interface {
	SaveRefreshToken(ctx context.Context, session models.Session) error
	GetSession(ctx context.Context, session models.Session) (*models.Session, error)
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
	_, err := d.conn.NamedQueryContext(ctx, insertOrUpdateSession, session)
	return utils.CheckError(err)
}

var getSessionQuery = `
	SELECT user_id, device_id, refresh_token, expires_at
	FROM sessions
	WHERE user_id = $1
	      AND device_id = $2
	      AND refresh_token = $3
	      AND to_timestamp(expires_at) > NOW()
`

func (d *database) GetSession(ctx context.Context, data models.Session) (*models.Session, error) {
	var session models.Session
	if err := d.conn.GetContext(ctx, &session, getSessionQuery, data.UserID, data.DeviceID, data.RefreshToken); err != nil {
		return nil, err
	}

	return &session, nil
}
