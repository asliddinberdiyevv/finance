package auth

import (
	"finance/internal/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// Very secret key
var jwtKey = []byte("my_secret_key") // TODO change key

var accessTokenDuration = time.Duration(30) * time.Minute // 30 minuts
var refreshTokenDuration = time.Duration(2) * time.Hour   // 2 hours

type Cliams struct {
	UserID models.UserID `json:"userID"`
	jwt.StandardClaims
}

// * Tokens is wrapper for access and refresh tokens
type Tokens struct {
	AccessToken           string `json:"access_token,omitempty"`
	AccessTokenExpiresAt  int64  `json:"expiresAt,omitempty"`
	RefreshToken          string `json:"refresh_token,omitempty"`
	RefreshTokenExpiresAt int64  `json:"-"` //! we will store this time in database with refersh token
}

// * IssueToken generate Access and Refresh token:
// * Currently we will generate only access token
func IssueToken(principal models.Principal) (*Tokens, error) {
	if principal.UserID == models.NilUserID {
		return nil, errors.New("invalid principal")
	}

	accessToken, accessTokenExpiresAt, err := GenerateToken(principal, accessTokenDuration)
	if err != nil {
		return nil, err
	}

	refreshToken, refreshTokenExpiresAt, err := GenerateToken(principal, refreshTokenDuration)
	if err != nil {
		return nil, err
	}

	tokens := Tokens{
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessTokenExpiresAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshTokenExpiresAt,
	}

	return &tokens, nil
}

func GenerateToken(principal models.Principal, duration time.Duration) (string, int64, error) {
	now := time.Now()

	// * Generate access token
	claims := &Cliams{
		UserID: principal.UserID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			ExpiresAt: now.Add(duration).Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", 0, err
	}

	return tokenString, claims.ExpiresAt, nil
}

func VerifyToken(token string) (*models.Principal, error) {
	cliams := &Cliams{}
	tkn, err := jwt.ParseWithClaims(token, cliams, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	principal := &models.Principal{
		UserID: cliams.UserID,
	}

	// We want to return principal even token invalid because we need to get UserID
	if !tkn.Valid {
		return principal, err
	}

	return principal, nil
}
