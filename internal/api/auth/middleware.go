package auth

import (
	"context"
	"finance/internal/api/utils"
	"finance/internal/models"
	"net/http"
	"strings"

	"github.com/pkg/errors"
)

type principalContextKeyType struct{}

var principalContextKey principalContextKeyType

func AuthorizationToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req, err := CheckToken(r)
		if err != nil {
			utils.WriteError(w, http.StatusUnauthorized, err.Error(), nil)
			return
		}

		next.ServeHTTP(w, req)
	})
}

func CheckToken(r *http.Request) (*http.Request, error) {
	// * extract token from header
	token, err := GetToken(r)
	if err != nil {
		return r, err
	}

	// ! allow to continue without token
	if token == "" {
		return r, nil
	}

	principal, err := VerifyToken(token)
	if err != nil {
		return r, err
	}

	return r.WithContext(WithPrincipalContext(r.Context(), *principal)), nil
}

// * Set principal in context to get it in API
func WithPrincipalContext(ctx context.Context, principal models.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func GetToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", nil
	}

	tokenParts := strings.SplitN(token, " ", 2)
	if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" || len(tokenParts[1]) == 0 {
		return "", errors.New("Authorization header format must be Bearer {token}")
	}

	return tokenParts[1], nil
}

func GetPrincipal(r *http.Request) models.Principal {
	if principal, ok := r.Context().Value(principalContextKey).(models.Principal); ok {
		return principal
	}
	return models.NilPrincipal
}
