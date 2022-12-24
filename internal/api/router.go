package api

import (
	"net/http"

	"github.com/asliddinberdiyevv/finance/internal/api/v1"
	"github.com/gorilla/mux"
)

func NewRouter() (http.Handler, error) {
	router := mux.NewRouter()
	router.HandleFunc("/version", v1.VersionHanler)

	return router, nil
}
