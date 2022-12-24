package main

import (
	"github.com/asliddinberdiyevv/finance/internal/api"
	"github.com/asliddinberdiyevv/finance/internal/config"
	"github.com/sirupsen/logrus"
	"net/http"
)

// Create server object and start
func main() {

	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithField("version", config.Version).Debug("Starting server.")

	// Create new router
	router, err := api.NewRouter()
	if err != nil {
		logrus.WithError(err).Fatal("Error building router")
	}

	const addr = "0.0.0.0:8088"
	server := http.Server{
		Handler: router,
		Addr:    addr,
	}

	// Starting server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Error("Server failed.")
	}
}
