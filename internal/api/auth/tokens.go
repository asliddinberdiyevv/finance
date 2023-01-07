package auth

import (
	"crypto/sha512"
	"finance/internal/models"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

// Very secret key
var jwtKey = []byte("my_secret_key") // TODO change key
var tokenDurationHours = 30 * 24     //30 days

type Tokens interface {
	IssueToken(principal models.Principal) (string, error)
	// Verify(token string) (*models.Principal, error)
}

type tokens struct {
	key             []byte
	duration        time.Duration
	beforeTolerance time.Duration
	signingMethod   jwt.SigningMethod
}

type Cliams struct {
	UserID models.UserID `json:"userID"`
	jwt.StandardClaims
}

func NewTokens() Tokens {
	hasher := sha512.New()

	if _, err := hasher.Write([]byte(jwtKey)); err != nil {
		panic(err)
	}

	tokenDuration := time.Duration(tokenDurationHours) * time.Hour

	return &tokens{
		key:             hasher.Sum(nil),
		duration:        tokenDuration,
		beforeTolerance: -2 * time.Minute,
		signingMethod:   jwt.SigningMethodHS512,
	}
}

func (t *tokens) IssueToken(principal models.Principal) (string, error) {
	if principal.UserID == models.NilUserID {
		return "", errors.New("invalid principal")
	}

	now := time.Now()
	claims := &Cliams{
		UserID: principal.UserID,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  now.Unix(),
			NotBefore: now.Add(t.beforeTolerance).Unix(),
			ExpiresAt: now.Add(t.duration).Unix(),
		},
	}

	token := jwt.NewWithClaims(t.signingMethod, claims)

	return token.SignedString(t.key)
}
