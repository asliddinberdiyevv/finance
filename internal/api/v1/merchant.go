package v1

import (
	"encoding/json"
	"finance/internal/api/auth"
	"finance/internal/database"
	"finance/internal/models"
	"finance/internal/utils"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// MerchantAPI - provides REST for Merchant
type MerchantAPI struct {
	DB database.Database
}

func SetMerchantAPI(db database.Database, router *mux.Router, permissons auth.Permissions) {
	api := MerchantAPI{
		DB: db,
	}

	apis := []API{
		/* ---------- MERCHANTS ---------- */
		NewAPI("/users/{userID}/merchants", "POST", api.Create, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/merchants", "GET", api.List, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/merchants/{merchantID}", "GET", api.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/merchants/{merchantID}", "PATCH", api.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/merchants/{merchantID}", "DELETE", api.Delete, auth.Admin, auth.MemberIsTarget),
	}

	for _, api := range apis {
		router.HandleFunc(api.Path, permissons.Wrap(api.Func, api.Permissions...)).Methods(api.Method)
	}
}

// POST - /users/{userID}/merchants
// Permission - MemberIsTarget
func (api *MerchantAPI) Create(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "merchant.go -> Create()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var merchant models.Merchant
	if err := json.NewDecoder(r.Body).Decode(&merchant); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	merchant.UserID = &userID

	if err := merchant.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Store role in database

	if err := api.DB.CreateMerchant(ctx, &merchant); err != nil {
		logger.WithError(err).Warn("Error creating merchant.")
		utils.WriteError(w, http.StatusInternalServerError, "Error creating merchant.", nil)
		return
	}

	logger.WithField("merchantID", merchant.ID).Info("Merchant created")
	utils.WriteJSON(w, http.StatusCreated, merchant)
}

// PATCH - /users/{userID}/merchants/{merchantID}
// Permission - MemberIsTarget
func (api *MerchantAPI) Update(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "merchant.go -> Update()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	merchantID := models.MerchantID(vars["merchantID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"merchant_id": merchantID,
	})

	// Decode parameters
	var merchantRequest models.Merchant
	if err := json.NewDecoder(r.Body).Decode(&merchantRequest); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	merchant, err := api.DB.GetMerchantByID(ctx, merchantID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting merchant.", http.StatusConflict)
		return
	}

	if merchantRequest.Name != nil || len(*merchantRequest.Name) != 0 {
		merchant.Name = merchantRequest.Name
	}

	if err := api.DB.UpdateMerchant(ctx, merchant); err != nil {
		logger.WithError(err).Warn("Error updating merchant.")
		utils.WriteError(w, http.StatusInternalServerError, "Error updating merchant.", nil)
		return
	}

	logger.Info("Merchant update")
	utils.WriteJSON(w, http.StatusOK, merchant)
}

// GET - /users/{userID}/merchants
// Permission - MemberIsTarget
func (api *MerchantAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "merchant.go -> List()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	ctx := r.Context()
	merchants, err := api.DB.ListMerchantByUserID(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting merchants.", http.StatusConflict)
		return
	}

	if merchants == nil {
		merchants = make([]*models.Merchant, 0)
	}

	logger.Info("Merchants returned")
	utils.WriteJSON(w, http.StatusOK, merchants)
}

// GET - /users/{userID}/merchants/{merchantID}
// Permission - MemberIsTarget
func (api *MerchantAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "merchant.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	merchantID := models.MerchantID(vars["merchantID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"merchant_id": merchantID,
	})

	ctx := r.Context()
	merchant, err := api.DB.GetMerchantByID(ctx, merchantID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting merchant.", http.StatusConflict)
		return
	}

	logger.Info("Merchant returned")
	utils.WriteJSON(w, http.StatusOK, merchant)
}

// DELETE - /users/{userID}/merchants/{merchantID}
// Permission - MemberIsTarget
func (api *MerchantAPI) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "merchant.go -> Delete()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	merchantID := models.MerchantID(vars["merchantID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"merchant_id": merchantID,
	})

	ctx := r.Context()
	deleted, err := api.DB.DeleteMerchant(ctx, merchantID)
	if err != nil {
		utils.ResponseErr(err, w, "Error deleting merchant.", http.StatusConflict)
		return
	}

	logger.Info("Merchant deleted")
	utils.WriteJSON(w, http.StatusOK, &ActDeleted{
		Deleted: deleted,
	})
}
