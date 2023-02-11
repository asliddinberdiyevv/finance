package main

import (
	"finance/internal/api"
	"finance/internal/config"
	"finance/internal/database"
	"fmt"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/namsral/flag"
	"github.com/sirupsen/logrus"
)

// Create server object and start
func main() {
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
	}

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

	var addr = ":" + os.Getenv("APP_PORT")
	server := http.Server{
		Handler: router,
		Addr:    addr,
	}

	// Starting server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Error("Server failed.")
	}
}

func New() {
	panic("unimplemented")
}
