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
type CategoryAPI struct {
	DB database.Database
}

func SetCategoryAPI(db database.Database, router *mux.Router, permissons auth.Permissions) {
	api := CategoryAPI{
		DB: db,
	}

	apis := []API{
		/* ---------- CATEGORIES ---------- */
		NewAPI("/users/{userID}/categories", "POST", api.Create, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories", "GET", api.List, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "GET", api.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "PATCH", api.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}/categories/{categoryID}", "DELETE", api.Delete, auth.Admin, auth.MemberIsTarget),
	}

	for _, api := range apis {
		router.HandleFunc(api.Path, permissons.Wrap(api.Func, api.Permissions...)).Methods(api.Method)
	}
}

// POST - /users/{userID}/categories
// Permission - MemberIsTarget
func (api *CategoryAPI) Create(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "category.go -> Create()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	category.UserID = &userID

	if err := category.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	// Store role in database
	if err := api.DB.CreateCategory(ctx, &category); err != nil {
		logger.WithError(err).Warn("Error creating category.")
		utils.WriteError(w, http.StatusInternalServerError, "Error creating category.", nil)
		return
	}

	logger.WithField("categoryID", category.ID).Info("Category created")
	utils.WriteJSON(w, http.StatusCreated, category)
}

// PATCH - /users/{userID}/categories/{categoryID}
// Permission - MemberIsTarget
func (api *CategoryAPI) Update(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "category.go -> Update()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	categoryID := models.CategoryID(vars["categoryID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"category_id": categoryID,
	})

	// Decode parameters
	var categoryRequest models.Category
	if err := json.NewDecoder(r.Body).Decode(&categoryRequest); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	category, err := api.DB.GetCategoryByID(ctx, categoryID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting category.", http.StatusConflict)
		return
	}

	if categoryRequest.ParentID != "" {
		category.ParentID = categoryRequest.ParentID
	}

	if categoryRequest.Name != nil || len(*categoryRequest.Name) != 0 {
		category.Name = categoryRequest.Name
	}

	if err := api.DB.UpdateCategory(ctx, category); err != nil {
		logger.WithError(err).Warn("Error updating category.")
		utils.WriteError(w, http.StatusInternalServerError, "Error updating category.", nil)
		return
	}

	logger.Info("Category update")
	utils.WriteJSON(w, http.StatusOK, category)
}

// GET - /users/{userID}/categories
// Permission - MemberIsTarget
func (api *CategoryAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "category.go -> List()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	ctx := r.Context()
	categories, err := api.DB.ListCategoryByUserID(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting categories.", http.StatusConflict)
		return
	}

	if categories == nil {
		categories = make([]*models.Category, 0)
	}

	logger.Info("Categories returned")
	utils.WriteJSON(w, http.StatusOK, categories)
}

// GET - /users/{userID}/categories/{categoryID}
// Permission - MemberIsTarget
func (api *CategoryAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "category.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	categoryID := models.CategoryID(vars["categoryID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"category_id": categoryID,
	})

	ctx := r.Context()
	category, err := api.DB.GetCategoryByID(ctx, categoryID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting category.", http.StatusConflict)
		return
	}

	logger.Info("Category returned")
	utils.WriteJSON(w, http.StatusOK, category)
}

// DELETE - /users/{userID}/categories/{categoryID}
// Permission - MemberIsTarget
func (api *CategoryAPI) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "category.go -> Delete()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	categoryID := models.CategoryID(vars["categoryID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":     userID,
		"principal":   principal,
		"category_id": categoryID,
	})

	ctx := r.Context()
	deleted, err := api.DB.DeleteCategory(ctx, categoryID)
	if err != nil {
		utils.ResponseErr(err, w, "Error deleting category.", http.StatusConflict)
		return
	}

	logger.Info("Category deleted")
	utils.WriteJSON(w, http.StatusOK, &ActDeleted{
		Deleted: deleted,
	})
}
