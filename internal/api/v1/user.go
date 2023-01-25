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
	DB database.Database // will represent all database interafaces
}

type UserParameters struct {
	models.User
	models.SessionData

	Password string `json:"password"`
}

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

	createdUser, err := api.DB.GetUserByID(ctx, &newUser.ID)
	if err != nil {
		utils.ResponseErr(err, w, "Error creating user.", http.StatusConflict)
		return
	}

	logger.WithField("userID", createdUser.ID).Info("User created")
	utils.WriteJSON(w, http.StatusCreated, createdUser)
	// api.writeTokenResponse(ctx, w, http.StatusCreated, createdUser, &userParameters.SessionData, true)
}

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
		utils.ResponseErr(err, w, "Invalid email or password.", http.StatusConflict)
		return
	}

	logger.WithField("userID", user.ID).Debug("user logged in")
	api.writeTokenResponse(ctx, w, http.StatusOK, user, &credential.SessionData, true)
}

func (api *UserAPI) Get(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> Get()")

	vars := mux.Vars(r)
	userID := models.UserID(vars["userID"])

	ctx := r.Context()
	user, err := api.DB.GetUserByID(ctx, &userID)
	if err != nil {
		utils.ResponseErr(err, w, "Error getting user.", http.StatusConflict)
		return
	}

	logger.WithField("userID", userID).Debug("Get user complete")
	utils.WriteJSON(w, http.StatusOK, user)
}

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
	user, err := api.DB.GetUserByID(ctx, &principal.UserID)
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
