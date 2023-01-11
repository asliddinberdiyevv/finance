package api

import (
	"finance/internal/api/auth"
	v1 "finance/internal/api/v1"
	"finance/internal/config"
	"finance/internal/database"
	"net/http"

	"github.com/gorilla/mux"
)

func NewRouter(db database.Database) (http.Handler, error) {
	router := mux.NewRouter()
	router.HandleFunc("/version", v1.VersionHanler)

	apiRouter := router.PathPrefix("/api/" + config.Version).Subrouter()

	userAPI := &v1.UserAPI{
		DB: db,
	}

	apiRouter.HandleFunc("/users", userAPI.Create).Methods("POST") // create user
	// apiRouter.HandleFunc("/users", userAPI.Create).Methods("GET") // list all users
	apiRouter.HandleFunc("/users/{userID}", userAPI.Get).Methods("GET") // get user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("PATCH") // update user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("DELETE") // delete user by id

	apiRouter.HandleFunc("/login", userAPI.Login).Methods("POST") // login in user

	router.Use(auth.AuthorizationToken)

	return router, nil
}
