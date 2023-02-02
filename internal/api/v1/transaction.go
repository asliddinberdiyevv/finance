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

// TransactionAPI - provides REST for Transaction
type TransactionAPI struct {
	DB database.Database
}

func SetTransactionAPI(db database.Database, router *mux.Router, permissons auth.Permissions) {
	api := TransactionAPI{
		DB: db,
	}

	apis := []API{
		/* ---------- TRANSACTION ---------- */
		NewAPI("/users/{userID}/transactions", "POST", api.Create, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/transactions", "GET", api.ListByUser, auth.Admin, auth.MemberIsTarget),
		NewAPI("/accounts/{accountID}/transactions", "GET", api.ListByAccount, auth.Admin, auth.MemberIsTarget),
		NewAPI("/categories/{categoryID}/transactions", "GET", api.ListByCategory, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/transactions/{transactionID}", "GET", api.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/transactions/{transactionID}", "PATCH", api.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/transactions/{transactionID}", "DELETE", api.Delete, auth.Admin, auth.MemberIsTarget),
	}

	for _, api := range apis {
		router.HandleFunc(api.Path, permissons.Wrap(api.Func, api.Permissions...)).Methods(api.Method)
	}
}

// POST - /users/{userID}/transactions
// Permission - MemberIsTarget
func (api *TransactionAPI) Create(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> Create()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var transaction models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transaction); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	transaction.UserID = &userID

	if err := transaction.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Store role in database

	if err := api.DB.CreateTransaction(ctx, &transaction); err != nil {
		logger.WithError(err).Warn("Error creating transaction.")
		utils.WriteError(w, http.StatusInternalServerError, "Error creating transaction.", nil)
		return
	}

	logger.WithField("transactionID", transaction.ID).Info("Transaction created")
	utils.WriteJSON(w, http.StatusCreated, transaction)
}

// PATCH - /users/{userID}/transactions/{transactionID}
// Permission - MemberIsTarget
func (api *TransactionAPI) Update(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> Update()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	transactionID := models.TransactionID(vars["transactionID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"principal":      principal,
		"transaction_id": transactionID,
	})

	// Decode parameters
	var transactionRequest models.Transaction
	if err := json.NewDecoder(r.Body).Decode(&transactionRequest); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	transaction, err := api.DB.GetTransactionByID(ctx, transactionID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting transaction.", http.StatusConflict)
		return
	}

	if transactionRequest.AccountID != nil || *transactionRequest.AccountID != models.NilAccountID {
		transaction.AccountID = transactionRequest.AccountID
	}

	if transactionRequest.CategoryID != nil || *transactionRequest.CategoryID != models.NilCategoryID {
		transaction.CategoryID = transactionRequest.CategoryID
	}

	if transactionRequest.Date != nil {
		transaction.Date = transactionRequest.Date
	}

	if transactionRequest.Type != nil || *transactionRequest.Type != "" {
		transaction.Type = transactionRequest.Type
	}

	if transactionRequest.Amount != nil {
		transaction.Amount = transactionRequest.Amount
	}

	if transactionRequest.Notes != nil {
		transaction.Notes = transactionRequest.Notes
	}

	if err := api.DB.UpdateTransaction(ctx, transaction); err != nil {
		logger.WithError(err).Warn("Error updating transaction.")
		utils.WriteError(w, http.StatusInternalServerError, "Error updating transaction.", nil)
		return
	}

	logger.Info("Transaction update")
	utils.WriteJSON(w, http.StatusOK, transaction)
}

// GET - /users/{userID}/transactions/{transactionID}
// Permission - MemberIsTarget
func (api *TransactionAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	transactionID := models.TransactionID(vars["transactionID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":        userID,
		"principal":      principal,
		"transaction_id": transactionID,
	})

	ctx := r.Context()
	transaction, err := api.DB.GetTransactionByID(ctx, transactionID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting transaction.", http.StatusConflict)
		return
	}

	logger.Info("Transaction returned")
	utils.WriteJSON(w, http.StatusOK, transaction)
}

// GET - /users/{userID}/transactions?from={from}&to={to}
// Permission - MemberIsTarget
func (api *TransactionAPI) ListByUser(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> ListByUser()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)
	query := r.URL.Query()
	
	from, err := utils.TimeParam(query, "from")
	if err != nil {
		utils.ResponseErr(err, w, "invaled from parameter.", http.StatusBadRequest)
		return
	}

	to, err := utils.TimeParam(query, "to")
	if err != nil {
		utils.ResponseErr(err, w, "invaled to parameter.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
		"from": from,
		"to": to,
	})

	ctx := r.Context()
	transactions, err := api.DB.ListTransactionByUserID(ctx, userID, from, to)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting transactions.", http.StatusConflict)
		return
	}

	if transactions == nil {
		transactions = make([]*models.Transaction, 0)
	}

	logger.Info("Transactions returned")
	utils.WriteJSON(w, http.StatusOK, transactions)
}

// GET - /accounts/{accountID}/transactions?from={from}&to={to}
// Permission - MemberIsTarget
func (api *TransactionAPI) ListByAccount(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> ListByAccount()")

	vars := mux.Vars(r)
	accountID := models.AccountID(vars["accountID"])
	principal := auth.GetPrincipal(r)
	query := r.URL.Query()
	
	from, err := utils.TimeParam(query, "from")
	if err != nil {
		utils.ResponseErr(err, w, "invaled from parameter.", http.StatusBadRequest)
		return
	}

	to, err := utils.TimeParam(query, "to")
	if err != nil {
		utils.ResponseErr(err, w, "invaled to parameter.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"account_id":   accountID,
		"principal": principal,
		"from": from,
		"to": to,
	})

	ctx := r.Context()
	transactions, err := api.DB.ListTransactionByAccountID(ctx, accountID, from, to)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting transactions.", http.StatusConflict)
		return
	}

	if transactions == nil {
		transactions = make([]*models.Transaction, 0)
	}

	logger.Info("Transactions returned")
	utils.WriteJSON(w, http.StatusOK, transactions)
}

// GET - /categories/{categoryID}/transactions?from={from}&to={to}
// Permission - MemberIsTarget
func (api *TransactionAPI) ListByCategory(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> ListByCategory()")

	vars := mux.Vars(r)
	categoryID := models.CategoryID(vars["categoryID"])
	principal := auth.GetPrincipal(r)
	query := r.URL.Query()
	
	from, err := utils.TimeParam(query, "from")
	if err != nil {
		utils.ResponseErr(err, w, "invaled from parameter.", http.StatusBadRequest)
		return
	}

	to, err := utils.TimeParam(query, "to")
	if err != nil {
		utils.ResponseErr(err, w, "invaled to parameter.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"category_id":   categoryID,
		"principal": principal,
		"from": from,
		"to": to,
	})

	ctx := r.Context()
	transactions, err := api.DB.ListTransactionByCategoryID(ctx, categoryID, from, to)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting transactions.", http.StatusConflict)
		return
	}

	if transactions == nil {
		transactions = make([]*models.Transaction, 0)
	}

	logger.Info("Transactions returned")
	utils.WriteJSON(w, http.StatusOK, transactions)
}

// DELETE - /users/{userID}/transactions/{transactionID}
// Permission - MemberIsTarget
func (api *TransactionAPI) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "transaction.go -> Delete()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	transactionID := models.TransactionID(vars["transactionID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"transaction_id": transactionID,
	})

	ctx := r.Context()
	deleted, err := api.DB.DeleteTransaction(ctx, transactionID)
	if err != nil {
		utils.ResponseErr(err, w, "Error deleting transaction.", http.StatusConflict)
		return
	}

	logger.Info("Transaction deleted")
	utils.WriteJSON(w, http.StatusOK, &ActDeleted{
		Deleted: deleted,
	})
}
