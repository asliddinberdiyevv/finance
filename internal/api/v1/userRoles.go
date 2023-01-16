package v1

import (
	"encoding/json"
	"finance/internal/api/auth"
	"finance/internal/utils"
	"finance/internal/models"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (api *UserAPI) GrantRole(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> GrantRole()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var userRole models.UserRole
	decodeErr := json.NewDecoder(r.Body).Decode(&userRole)
	utils.ResponseErrWithMap(decodeErr, w, "Could not decode parametrs.", http.StatusBadRequest)

	ctx := r.Context()
	// Store role in database
	if err := api.DB.GrantRole(ctx, userID, userRole.Role); err != nil {
		logger.WithError(err).Warn("Error granting role.")
		utils.WriteError(w, http.StatusInternalServerError, "Error granting role.", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, &ActCreated{
		Created: true,
	})
}

func (api *UserAPI) RevokeRole(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> RevokeRole()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var userRole models.UserRole
	if err := json.NewDecoder(r.Body).Decode(&userRole); err != nil {
		logger.WithError(err).Warn("could not decode parametrs")
		utils.WriteError(w, http.StatusBadRequest, "could not decode parametrs", map[string]string{
			"error": err.Error(),
		})
		return
	}

	ctx := r.Context()
	// Store role in database
	if err := api.DB.RevokeRole(ctx, userID, userRole.Role); err != nil {
		logger.WithError(err).Warn("Error revoking role.")
		utils.WriteError(w, http.StatusInternalServerError, "Error revoking role.", nil)
		return
	}
	utils.WriteJSON(w, http.StatusCreated, &ActDeleted{
		Deleted: true,
	})
}

func (api *UserAPI) GetRoleList(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> GetRoleList()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	ctx := r.Context()
	// Store role in database
	roles, err := api.DB.GetRolesByUser(ctx, userID)
	if err != nil {
		logger.WithError(err).Warn("Error getting roles.")
		utils.WriteError(w, http.StatusInternalServerError, "Error getting roles.", nil)
		return
	}

	if roles == nil {
		roles = make([]*models.UserRole, 0)
	}

	utils.WriteJSON(w, http.StatusOK, &roles)
}
