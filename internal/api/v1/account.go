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

// AccountAPI - provides REST for Account
type AccountAPI struct {
	DB database.Database
}

// POST - /users/{userID}/accounts
// Permission - MemberIsTarget
func (api *AccountAPI) Create(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "account.go -> Create()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var account models.Account
	if err := json.NewDecoder(r.Body).Decode(&account); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	account.UserID = &userID

	if err := account.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Store role in database
	if err := api.DB.CreateAccount(ctx, &account); err != nil {
		logger.WithError(err).Warn("Error creating account.")
		utils.WriteError(w, http.StatusInternalServerError, "Error creating account.", nil)
		return
	}

	logger.WithField("accountID", account.ID).Info("Account created")
	utils.WriteJSON(w, http.StatusCreated, account)
}

// PATCH - /users/{userID}/accounts/{accountID}
// Permission - MemberIsTarget
func (api *AccountAPI) Update(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "account.go -> Update()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	accountID := models.AccountID(vars["accountID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"principal":  principal,
		"account_id": accountID,
	})

	// Decode parameters
	var accountRequest models.Account
	if err := json.NewDecoder(r.Body).Decode(&accountRequest); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	account, err := api.DB.GetAccountByID(ctx, accountID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting account.", http.StatusConflict)
		return
	}

	if accountRequest.Name != nil || len(*accountRequest.Name) != 0 {
		account.Name = accountRequest.Name
	}
	if accountRequest.Type != nil || len(*accountRequest.Type) != 0 {
		account.Type = accountRequest.Type
	}
	if accountRequest.StartBalance != nil {
		account.StartBalance = accountRequest.StartBalance
	}
	if accountRequest.Currency != nil || len(*accountRequest.Currency) != 0 {
		account.Currency = accountRequest.Currency
	}

	if err := api.DB.UpdateAccount(ctx, account); err != nil {
		logger.WithError(err).Warn("Error updating account.")
		utils.WriteError(w, http.StatusInternalServerError, "Error updating account.", nil)
		return
	}

	logger.Info("Account update")
	utils.WriteJSON(w, http.StatusOK, account)
}

// GET - /users/{userID}/accounts
// Permission - MemberIsTarget
func (api *AccountAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "account.go -> List()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	ctx := r.Context()
	accounts, err := api.DB.ListAccountByUserID(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting accounts.", http.StatusConflict)
		return
	}
	if accounts == nil {
		accounts = make([]*models.Account, 0)
	}

	logger.Info("Accounts returned")
	utils.WriteJSON(w, http.StatusOK, accounts)
}

// GET - /users/{userID}/accounts/{accountID}
// Permission - MemberIsTarget
func (api *AccountAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "account.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	accountID := models.AccountID(vars["accountID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"principal":  principal,
		"account_id": accountID,
	})

	ctx := r.Context()
	account, err := api.DB.GetAccountByID(ctx, accountID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting account.", http.StatusConflict)
		return
	}

	logger.Info("Account returned")
	utils.WriteJSON(w, http.StatusOK, account)
}

// DELETE - /users/{userID}/accounts/{accountID}
// Permission - MemberIsTarget
func (api *AccountAPI) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "account.go -> Delete()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	accountID := models.AccountID(vars["accountID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":    userID,
		"principal":  principal,
		"account_id": accountID,
	})

	ctx := r.Context()
	deleted, err := api.DB.DeleteAccount(ctx, accountID)
	if err != nil {
		utils.ResponseErr(err, w, "Error deleting account.", http.StatusConflict)
		return
	}

	logger.Info("Account deleted")
	utils.WriteJSON(w, http.StatusOK, &ActDeleted{
		Deleted: deleted,
	})
}
