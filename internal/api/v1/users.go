package v1

import (
	"context"
	"encoding/json"
	"finance/internal/api/auth"
	"finance/internal/api/utils"

	"finance/internal/database"
	"finance/internal/models"
	"net/http"

	"github.com/sirupsen/logrus"
)

// UserAPI - provides REST for Users
type UserAPI struct {
	DB database.Database // will represent all database interafaces

	Tokens auth.Tokens
}

type UserParameters struct {
	models.User
	Password string `json:"password"`
}

func (api *UserAPI) Create(w http.ResponseWriter, r *http.Request) {
	// show function name is logs to track error faster
	logger := logrus.WithField("func", "users.go -> Create()")

	// Load parameters
	var userParameters UserParameters
	if err := json.NewDecoder(r.Body).Decode(&userParameters); err != nil {
		logger.WithError(err).Warn("could not decode parametrs")
		utils.WriteError(w, http.StatusBadRequest, "could not decode parametrs", map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"email": *userParameters.Email,
	})

	if err := userParameters.Verify(); err != nil {
		logger.WithError(err).Warn("Not all fields found.")
		utils.WriteError(w, http.StatusInternalServerError, "Not all fields found", map[string]string{
			"error": err.Error(),
		})
		return
	}

	hashed, err := models.HashPassword(userParameters.Password)
	if err != nil {
		logger.WithError(err).Warn("Could not hash password.")
		utils.WriteError(w, http.StatusInternalServerError, "Could not hash password", nil)
		return
	}

	newUser := &models.User{
		Email:        userParameters.Email,
		PasswordHash: &hashed,
	}

	ctx := r.Context()

	if err := api.DB.CreateUser(ctx, newUser); err == database.ErrUserExists {
		logger.WithError(err).Warn("User already exists")
		utils.WriteError(w, http.StatusConflict, "User already exists", nil)
	} else if err != nil {
		logger.WithError(err).Warn("Error creating user")
		utils.WriteError(w, http.StatusConflict, "Error creating user", nil)
	}

	createdUser, err := api.DB.GetUserByID(ctx, &newUser.ID)
	if err != nil {
		logger.WithError(err).Warn("Error creating user")
		utils.WriteError(w, http.StatusConflict, "Error creating user", nil)
	}

	logger.Info("User created")

	// utils.WriteJSON(w, http.StatusCreated, createdUser)
	api.writeTokenResponse(ctx, w, http.StatusCreated, createdUser, true)
}

func (api *UserAPI) Login(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("func", "users.go -> Login()")

	var credentials models.Credentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		logger.WithError(err).Warn("could not decode parametrs")
		utils.WriteError(w, http.StatusBadRequest, "could not decode parametrs", map[string]string{
			"error": err.Error(),
		})
		return
	}

	logger = logger.WithFields(logrus.Fields{
		"email": credentials.Email,
	})

	ctx := r.Context()
	user, err := api.DB.GetUserByEmail(ctx, credentials.Email)
	if err != nil {
		logger.WithError(err).Warn("Error logging in")
		utils.WriteError(w, http.StatusConflict, "invalid email or password", nil)
		return
	}

	// Checking if password is correct
	if err := user.CheckPassword(credentials.Password); err != nil {
		logger.WithError(err).Warn("Error logging in")
		utils.WriteError(w, http.StatusConflict, "invalid email or password", nil)
		return
	}

	logger.WithField("userID", user.ID).Debug("user logged in")

	// utils.WriteJSON(w, http.StatusOK, user)
	api.writeTokenResponse(ctx, w, http.StatusOK, user, true)
}

type TokenResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user,omitempty"`
}

func (api *UserAPI) writeTokenResponse(ctx context.Context, w http.ResponseWriter, status int, user *models.User, cookie bool) {
	// Issue token:
	token, err := api.Tokens.IssueToken(models.Principal{UserID: user.ID})
	if err != nil {
		logrus.WithError(err).Warn("Error issuing token.")
		utils.WriteError(w, http.StatusUnauthorized, "Error issuing token", nil)
		return
	}

	// Write token response:
	tokenResponse := TokenResponse{
		Token: token,
		User:  user,
	}

	// if cookie {
		// later
	// }

	utils.WriteJSON(w, status, tokenResponse)

}
