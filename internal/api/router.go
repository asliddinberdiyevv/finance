package api

import (
	"finance/internal/api/auth"
	"finance/internal/api/v1"
	"finance/internal/config"
	"finance/internal/database"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(db database.Database) (http.Handler, error) {
	permissons := auth.NewPermissions(db)

	router := mux.NewRouter()
	router.HandleFunc("/version", v1.VersionHanler)

	apiRouter := router.PathPrefix("/api/" + config.Version).Subrouter()

	userAPI := &v1.UserAPI{
		DB: db,
	}

	/* ---------- USERS ---------- */
	apiRouter.HandleFunc("/users", userAPI.Create).Methods("POST") // create user
	// apiRouter.HandleFunc("/users", userAPI.Create).Methods("GET") // get all users
	apiRouter.HandleFunc("/users/{userID}", userAPI.Get).Methods("GET") // get user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("PATCH") // update user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("DELETE") // delete user by id

	/* ---------- LOGIN ---------- */
	apiRouter.HandleFunc("/login", userAPI.Login).Methods("POST")

	/* ---------- TOKENS ---------- */
	apiRouter.HandleFunc("/refresh", permissons.Wrap(userAPI.RefreshToken, auth.Member)).Methods("POST")

	/* ---------- ROLES ---------- */
	apiRouter.HandleFunc("/users/{userID}/roles", userAPI.GrantRole).Methods("POST")                               // Create role
	apiRouter.HandleFunc("/users/{userID}/roles", permissons.Wrap(userAPI.GetRoleList, auth.Admin)).Methods("GET") // Get all roles
	apiRouter.HandleFunc("/users/{userID}/roles", userAPI.RevokeRole).Methods("DELETE")                            // Delete role

	/* ---------- MIDDLEWARE ---------- */
	router.Use(auth.AuthorizationToken)

	return router, nil
}
