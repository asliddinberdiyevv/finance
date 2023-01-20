package main

import (
	"finance/internal/api"
	"finance/internal/config"
	"finance/internal/database"
	"net/http"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
)

// Create server object and start
func main() {
	flag.Parse()

	logrus.SetLevel(logrus.DebugLevel)
	logrus.WithField("version", config.Version).Debug("Starting server.")

	// tokens := auth.NewTokens()

	// Create new database
	db, err := database.New()
	if err != nil {
		logrus.WithError(err).Fatal("Error verifying database.")
	}
	logrus.Debug("Database is ready to use.")

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
