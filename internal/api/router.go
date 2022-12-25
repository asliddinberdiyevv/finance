package api

import (
	"net/http"

	"finance/internal/api/v1"
	"finance/internal/database"

	"github.com/gorilla/mux"
)

func NewRouter(db database.Database) (http.Handler, error) {
	router := mux.NewRouter()
	router.HandleFunc("/version", v1.VersionHanler)

	return router, nil
}
