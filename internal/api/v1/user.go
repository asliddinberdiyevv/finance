package v1

import (
	"context"
	"encoding/json"
	"finance/internal/api/auth"
	"finance/internal/utils"

	"finance/internal/database"
	"finance/internal/models"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// UserAPI - provides REST for Users
type UserAPI struct {
	DB database.Database
}

func SetUserAPI(db database.Database, router *mux.Router, permissons auth.Permissions) {
	api := UserAPI{
		DB: db,
	}

	apis := []API{
		/* ---------- USERS ---------- */
		NewAPI("/users", "POST", api.Create, auth.Any),
		NewAPI("/users", "GET", api.List, auth.Admin),
		NewAPI("/users/{userID}", "GET", api.Get, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}", "PATCH", api.Update, auth.Admin, auth.MemberIsTarget),
		NewAPI("/users/{userID}", "DELETE", api.Delete, auth.Admin, auth.MemberIsTarget),

		/* ---------- LOGIN ---------- */
		NewAPI("/login", "POST", api.Login, auth.Any),

		/* ---------- TOKENS ---------- */
		NewAPI("/refresh", "POST", api.RefreshToken, auth.Member),
	}

	for _, api := range apis {
		router.HandleFunc(api.Path, permissons.Wrap(api.Func, api.Permissions...)).Methods(api.Method)
	}
}

type UserParameters struct {
	models.User
	models.SessionData

	Password string `json:"password"`
}

/* ---------- USERS ---------- */

// POST - /users
// Permission - Any
func (api *UserAPI) Create(w http.ResponseWriter, r *http.Request) {
	// show function name is logs to track error faster
	logger := logrus.WithField("func", "users.go -> Create()")

	// Load parameters
	var userParameters UserParameters
	if err := json.NewDecoder(r.Body).Decode(&userParameters); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"email": *userParameters.Email,
	})

	if err := userParameters.User.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}
	if err := userParameters.SessionData.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	hashed, err := models.HashPassword(userParameters.Password)
	if err != nil {
		utils.ResponseErr(err, w, "Could not hash password.", http.StatusInternalServerError)
		return
	}

	newUser := &models.User{
		Email:        userParameters.Email,
		PasswordHash: &hashed,
	}

	ctx := r.Context()

	if err := api.DB.CreateUser(ctx, newUser); err == database.ErrUserExists {
		utils.ResponseErr(err, w, "User already exists.", http.StatusConflict)
		return
	} else if err != nil {
		utils.ResponseErr(err, w, "Error creating user.", http.StatusConflict)
		return
	}

	createdUser, err := api.DB.GetUserByID(ctx, newUser.ID)
	if err != nil {
		utils.ResponseErr(err, w, "Error creating user.", http.StatusConflict)
		return
	}

	logger.WithField("userID", createdUser.ID).Info("User created")
	utils.WriteJSON(w, http.StatusCreated, createdUser)
}

// GET - /users
// Permission - Admin
func (api *UserAPI) List(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "user.go -> List()")
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"principal": principal,
	})

	ctx := r.Context()
	users, err := api.DB.ListUsers(ctx)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting users.", http.StatusConflict)
		return
	}

	if users == nil {
		users = make([]*models.User, 0)
	}

	logger.Info("Users returned")
	utils.WriteJSON(w, http.StatusOK, users)
}

// GET - /users/{userID}
// Permission - Admin, MemberIsTarget
func (api *UserAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])

	ctx := r.Context()
	user, err := api.DB.GetUserByID(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting user.", http.StatusConflict)
		return
	}

	logger.WithField("userID", userID).Debug("Get user complete")
	utils.WriteJSON(w, http.StatusOK, user)
}

// PATCH - /users/{userID}
// Permission - Admin, MemberIsTarget
func (api *UserAPI) Update(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "user.go -> Update()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	// Decode parameters
	var userRequest UserParameters
	if err := json.NewDecoder(r.Body).Decode(&userRequest); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	user, err := api.DB.GetUserByID(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting user.", http.StatusConflict)
		return
	}

	if len(userRequest.Password) != 0 {
		if err := user.SetPassword(userRequest.Password); err != nil {
			utils.ResponseErr(err, w, "Error setting password.", http.StatusInternalServerError)
			return
		}
	}

	if err := api.DB.UpdateUser(ctx, user); err != nil {
		logger.WithError(err).Warn("Error updating user.")
		utils.WriteError(w, http.StatusInternalServerError, "Error updating user.", nil)
		return
	}

	logger.Info("User update")
	utils.WriteJSON(w, http.StatusOK, user)
}

// DELETE - /users/{userID}
// Permission - Admin, MemberIsTarget
func (api *UserAPI) Delete(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "user.go -> Delete()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])
	principal := auth.GetPrincipal(r)

	logger = logger.WithFields(logrus.Fields{
		"user_id":   userID,
		"principal": principal,
	})

	ctx := r.Context()
	deleted, err := api.DB.DeleteUser(ctx, userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error deleting user.", http.StatusConflict)
		return
	}

	logger.Info("User deleted")
	utils.WriteJSON(w, http.StatusOK, &ActDeleted{
		Deleted: deleted,
	})
}

/* ---------- LOGIN ---------- */
func (api *UserAPI) Login(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> Login()")

	var credential models.Credential
	if err := json.NewDecoder(r.Body).Decode(&credential); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"email": credential.Email,
	})

	ctx := r.Context()
	user, err := api.DB.GetUserByEmail(ctx, credential.Email)
	if err != nil {
		utils.ResponseErr(err, w, "Invalid email or password.", http.StatusConflict)
		return
	}
	if err := credential.SessionData.Verify(); err != nil {
		utils.ResponseErrWithMap(err, w, "Not all fields found.", http.StatusBadRequest)
		return
	}

	// Checking if password is correct
	if err := user.CheckPassword(credential.Password); err != nil {
		utils.ResponseErr(err, w, "Invalid email or password.", http.StatusUnauthorized)
		return
	}

	logger.WithField("userID", user.ID).Debug("user logged in")
	api.writeTokenResponse(ctx, w, http.StatusOK, user, &credential.SessionData, true)
}

/* ---------- TOKEN ---------- */
// RefreshTokenRequest - Data user send to get new access and refresh tokens.
type RefreshTokenRequest struct {
	RefreshToken string          `json:"refresh_token"`
	DeviceID     models.DeviceID `json:"device_id"`
}

func (api *UserAPI) RefreshToken(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> RefreshToken()")

	// TODO move this part to seperate function
	var request RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		utils.ResponseErrWithMap(err, w, "Could not decode parametrs.", http.StatusBadRequest)
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"device_id": request.DeviceID,
	})

	principal, err := auth.VerifyToken(request.RefreshToken)
	if err != nil {
		utils.ResponseErr(err, w, "Error verifing refresh token.", http.StatusUnauthorized)
		return
	}

	// * If token is valid we need to check if combination UserID - DeviceID - Refresh token exists in database
	data := models.Session{
		UserID:       principal.UserID,
		DeviceID:     request.DeviceID,
		RefreshToken: request.RefreshToken,
	}

	ctx := r.Context()
	session, err := api.DB.GetSession(ctx, data)
	if err != nil || session == nil {
		utils.ResponseErr(err, w, "Error session not exists.", http.StatusUnauthorized)
		return
	}

	// if session exists and valid we generate new access and refresh tokens.
	logger.WithField("user_id", principal.UserID).Debug("Refresh token")

	// Check if user exists
	user, err := api.DB.GetUserByID(ctx, principal.UserID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting user.", http.StatusConflict)
		return
	}

	api.writeTokenResponse(ctx, w, http.StatusOK, user, &models.SessionData{DeviceID: request.DeviceID}, true)
}

type TokenResponse struct {
	Tokens *auth.Tokens `json:"tokens,omitempty"` //this will insert all tokens struct fields
	User   *models.User `json:"user,omitempty"`
}

// writeTokenResponse - Generate Access and refresh token are return them to  user. Refresh token is stored in database as session
func (api *UserAPI) writeTokenResponse(ctx context.Context, w http.ResponseWriter, status int, user *models.User, sessionData *models.SessionData, cookie bool) {
	// Issue token:
	// TODO: add user role to Principal
	tokens, err := auth.IssueToken(models.Principal{UserID: user.ID})
	if err != nil || tokens == nil {
		utils.ResponseErr(err, w, "Error issuing token.", http.StatusUnauthorized)
		return
	}

	session := models.Session{
		UserID:       user.ID,
		DeviceID:     sessionData.DeviceID,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.RefreshTokenExpiresAt,
	}

	if err := api.DB.SaveRefreshToken(ctx, session); err != nil {
		utils.ResponseErr(err, w, "Error issuing token.", http.StatusUnauthorized)
		return
	}

	// Write token response:
	tokenResponse := TokenResponse{
		Tokens: tokens,
		User:   user,
	}

	// if cookie {
	// later
	// }

	utils.WriteJSON(w, status, tokenResponse)
}
