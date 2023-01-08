package auth

import (
	"finance/internal/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// Very secret key
var jwtKey = []byte("my_secret_key") // TODO change key

var accessTokenDuration = time.Duration(30) * time.Minute // 30 min
//! var refreshTokenDuration = 30 * 24 // 30 days

//? type Tokens interface {
//? 	IssueToken(principal models.Principal) (string, error)
//?  Verify(token string) (*models.Principal, error)
//? }

//? type tokens struct {
//? 	key           []byte
//? 	duration      time.Duration
//? 	signingMethod jwt.SigningMethod
//? }

type Cliams struct {
	UserID models.UserID `json:"userID"`
	jwt.StandardClaims
}

//? func NewTokens() Tokens {
//? 	tokenDuration := time.Duration(tokenDurationHours) * time.Hour

//?	return &tokens{
//?	key:           jwtKey,
//?		duration:      tokenDuration,
//?		signingMethod: jwt.SigningMethodHS512,
//?	}
//?}

// * Tokens is wrapper for access and refresh tokens
type Tokens struct {
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// * IssueToken generate Access and Refresh token:
// * Currently we will generate only access token
func IssueToken(principal models.Principal) (*Tokens, error) {
	if principal.UserID == models.NilUserID {
		return nil, errors.New("invalid principal")
	}

	now := time.Now()
	claims := &Cliams{
		UserID: principal.UserID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(accessTokenDuration).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	accessToken, err := token.SignedString(jwtKey)
	if err != nil {
		return nil, err
	}

	tokens := Tokens{
		AccessToken: accessToken,
	}

	return &tokens, nil
}

func VerifyToken(accessToken string) (*models.Principal, error) {
	cliams := &Cliams{}

	tkn, err := jwt.ParseWithClaims(accessToken, cliams, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			return nil, err
		}
		return nil, err
	}

	if !tkn.Valid {
		return nil, err
	}

	principal := &models.Principal{
		UserID: models.UserID(cliams.Id),
	}

	return principal, nil
}
