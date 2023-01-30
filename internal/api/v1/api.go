package v1

import (
	"finance/internal/api/auth"
	"net/http"
)

type API struct {
	Path        string
	Method      string
	Func        http.HandlerFunc
	Permissions []auth.PermissionTypes
}

func NewAPI(path string, method string, handlerFunc http.HandlerFunc, permissionTypes ...auth.PermissionTypes) API {
	return API{path, method, handlerFunc, permissionTypes}
}
