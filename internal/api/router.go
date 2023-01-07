package api

import (
	"net/http"

	"finance/internal/api/auth"
	"finance/internal/api/v1"
	"finance/internal/database"

	"github.com/gorilla/mux"
)

func NewRouter(db database.Database, tokens auth.Tokens) (http.Handler, error) {
	router := mux.NewRouter()
	router.HandleFunc("/version", v1.VersionHanler)

	apiRouter := router.PathPrefix("/api/v1").Subrouter()

	userAPI := &v1.UserAPI{
		DB:     db,
		Tokens: tokens,
	}

	apiRouter.HandleFunc("/users", userAPI.Create).Methods("POST") // create user
	// apiRouter.HandleFunc("/users", userAPI.Create).Methods("GET") // list all users
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("GET") // get user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("PATCH") // update user by id
	// apiRouter.HandleFunc("/users/{userID}", userAPI.Create).Methods("DELETE") // delete user by id

	apiRouter.HandleFunc("/login", userAPI.Login).Methods("POST") // login in user

	return router, nil
}
