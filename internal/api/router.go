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

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/version", v1.VersionHanler)
	apiRouter := router.PathPrefix("/api/" + config.Version).Subrouter()

	/* ---------- ROUTES ---------- */
	v1.SetUserAPI(db, apiRouter, permissons)
	v1.SetRoleApi(db, apiRouter, permissons)
	v1.SetCategoryAPI(db, apiRouter, permissons)
	v1.SetAccountAPI(db, apiRouter, permissons)
	v1.SetMerchantAPI(db, apiRouter, permissons)
	v1.SetTransactionAPI(db, apiRouter, permissons)

	/* ---------- MIDDLEWARE ---------- */
	router.Use(auth.AuthorizationToken)

	return router, nil
}
