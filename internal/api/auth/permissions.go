package auth

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"finance/internal/utils"
	"finance/internal/database"
	"finance/internal/models"

	"github.com/bluele/gcache"
	"github.com/gorilla/mux"
)

type Permissions interface {
	Wrap(next http.HandlerFunc, permissionTypes ...PermissionTypes) http.HandlerFunc
	Check(r *http.Request, permissionTypes ...PermissionTypes) bool
}

type permissions struct {
	DB    database.Database
	cache gcache.Cache
}

func NewPermissions(db database.Database) Permissions {
	p := &permissions{
		DB: db,
	}

	p.cache = gcache.New(200).
		LRU().
		LoaderExpireFunc(func(key interface{}) (interface{}, *time.Duration, error) {
			userID := key.(models.UserID)
			roles, err := p.DB.GetRolesByUser(context.Background(), userID)
			if err != nil {
				return nil, nil, err
			}
			expire := 1 * time.Minute
			return roles, &expire, nil
		}).
		Build()

	return p
}

// Get user's roles from cache (if we wont have roles in cache it will get it from database)
func (p *permissions) getRoles(userID models.UserID) ([]*models.UserRole, error) {
	roles, err := p.cache.Get(userID)
	if err != nil {
		return nil, err
	}
	return roles.([]*models.UserRole), nil
}

func (p *permissions) withRoles(principal models.Principal, roleFunc func([]*models.UserRole) bool) (bool, error) {
	if principal.UserID == models.NilUserID {
		return false, nil
	}

	// Load roles
	roles, err := p.getRoles(principal.UserID)
	if err != nil {
		return false, err
	}
	return roleFunc(roles), nil
}

// We need to see if we have principal on Request in this point...
func (p *permissions) Wrap(next http.HandlerFunc, permissionTypes ...PermissionTypes) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowed := p.Check(r, permissionTypes...); allowed {
			utils.WriteError(w, http.StatusUnauthorized, "permission denied", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// The idea is to return TRUE if one of permission types matches.
// For example if permission type is Admin and MemberIsTarget
// Admin can edit any userso if user has Admin role we don't care, admin don't match MemberIsTarget permission
func (p *permissions) Check(r *http.Request, permissionTypes ...PermissionTypes) bool {
	principal := GetPrincipal(r)
	for _, permissionType := range permissionTypes {
		fmt.Println(permissionType)
		switch permissionType {
		case Admin:
			if allowed, _ := p.withRoles(principal, adminOnly); allowed {
				return true
			}
		case Member:
			if allowed := member(principal); allowed {
				return true
			}
		case MemberIsTarget:
			targetUserID := models.UserID(mux.Vars(r)["userID"])
			if allowed := memberIsTarget(targetUserID, principal); allowed {
				return true
			}
		}
	}
	return false
}
