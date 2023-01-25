package api

import (
	"finance/internal/api/auth"
	v1 "finance/internal/api/v1"
	"finance/internal/config"
	"finance/internal/database"
	"net/http"

	"github.com/gorilla/mux"
)

type API struct {
	Path        string
	Method      string
	Func        http.HandlerFunc
	Permissions []auth.PermissionTypes
}

func NewRouter(db database.Database) (http.Handler, error) {
	permissons := auth.NewPermissions(db)

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/version", v1.VersionHanler)

	apiRouter := router.PathPrefix("/api/" + config.Version).Subrouter()

	userAPI := &v1.UserAPI{
		DB: db,
	}
	accountAPI := &v1.AccountAPI{
		DB: db,
	}
	categoryAPI := &v1.CategoryAPI{
		DB: db,
	}

	apis := []API{
		/* ---------- LOGIN ---------- */
		NewAPI("/login", "POST", userAPI.Login, auth.Any),

		/* ---------- TOKENS ---------- */
		NewAPI("/refresh", "POST", userAPI.RefreshToken, auth.Member),

		/* ---------- USERS ---------- */
		NewAPI("/users", "POST", userAPI.Create, auth.Member),
		// NewAPI("/users", "GET", userAPI.Create, auth.Admin),
		NewAPI("/users/{userID}", "GET", userAPI.Get, auth.Admin, auth.MemberIsTarget),
		// NewAPI("/users/{userID}", "PATCH", userAPI.Get, auth.Admin, auth.MemberIsTarget),
		// NewAPI("/users/{userID}", "DELETE", userAPI.Get, auth.Admin),

		/* ---------- ROLES ---------- */
		NewAPI("/users/{userID}/roles", "POST", userAPI.GrantRole, auth.Admin),
		NewAPI("/users/{userID}/roles", "GET", userAPI.GetRoleList, auth.Admin),
		NewAPI("/users/{userID}/roles", "DELETE", userAPI.RevokeRole, auth.Admin),

		/* ---------- ACCOUNTS ---------- */
		NewAPI("/users/{userID}/accounts", "POST", accountAPI.Create, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/accounts", "GET", accountAPI.List, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/accounts/{accountID}", "GET", accountAPI.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/accounts/{accountID}", "PATCH", accountAPI.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/accounts/{accountID}", "DELETE", accountAPI.Delete, auth.Admin, auth.MemberIsTarget),

		/* ---------- CATEGORIES ---------- */
		NewAPI("/users/{userID}/categories", "POST", categoryAPI.Create, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories", "GET", categoryAPI.List, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "GET", categoryAPI.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "PATCH", categoryAPI.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "DELETE", categoryAPI.Delete, auth.Admin, auth.MemberIsTarget),
	}

	for _, api := range apis {
		apiRouter.HandleFunc(api.Path, permissons.Wrap(api.Func, api.Permissions...)).Methods(api.Method)
	}

	/* ---------- MIDDLEWARE ---------- */
	router.Use(auth.AuthorizationToken)

	return router, nil
}

func NewAPI(path string, method string, handlerFunc http.HandlerFunc, permissionTypes ...auth.PermissionTypes) API {
	return API{path, method, handlerFunc, permissionTypes}
}
