package main

import (
	"finance/internal/api"
	"finance/internal/config"
	"finance/internal/database"
	"net/http"

	"github.com/sirupsen/logrus"
  _	"github.com/lib/pq"
)

// Create server object and start
func main() {

	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithField("version", config.Version).Debug("Starting server.")

	// Create new database
	db, err := database.New()
	if err != nil {
		logrus.WithError(err).Fatal("Error verifying database.")
	} 

	// Create new router
	router, err := api.NewRouter(db)
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
