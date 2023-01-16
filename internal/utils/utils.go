package utils

import (
	"encoding/json"
	"net/http"

	"github.com/sirupsen/logrus"
)

// GenerecError - represent error structure for generic error (we need this to make all our error response the same)
type GenerecError struct {
	Code  int         `json:"code"`
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}

// WriteError returns a JSON error message and HTTP status code.
func WriteError(w http.ResponseWriter, code int, message string, data interface{}) {
	response := GenerecError{
		Error: message,
		Code:  code,
		Data:  data,
	}

	WriteJSON(w, code, response)

}

// WriteJSON returns a JSON data and HTTP status code
func WriteJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		logrus.WithError(err).Warn("Error writing response.")
	}
}

// CheckError return error
func CheckError(err error) error {
	if err != nil {
		return err
	}
	return nil
}

// ResponseErr
func ResponseErr(err error, w http.ResponseWriter, msg string, status int) {
	logrus.WithError(err).Warn(msg)
	WriteError(w, status, msg, nil)
}

// ResponseErrWithMap
func ResponseErrWithMap(err error, w http.ResponseWriter, msg string, status int) {
	logrus.WithError(err).Warn(msg)
	WriteError(w, status, msg, map[string]string{
		"error": err.Error(),
	})
}
