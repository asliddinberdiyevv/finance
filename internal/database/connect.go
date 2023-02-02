package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
	"github.com/namsral/flag"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	databaseTimeout = flag.Int64("database-timeout-ms", 5000, "")
)

// Connect creates a new database connection
func Connect() (*sqlx.DB, error) {

	if err := godotenv.Load(); err != nil {
		fmt.Println(err)
	}

	username := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	var databaseURL = flag.String("database-url", "postgres://" + username +":"+ password +"@"+ dbHost +":"+ dbPort +"/"+ 
	dbName +"?sslmode=disable", "Database URL.")

	// Connect to database:
	dbURL := *databaseURL

	logrus.Debug("Connecting to database.")
	conn, err := sqlx.Open("postgres", dbURL)
	if err != nil {
		return nil, errors.Wrap(err, "Could not connect to database")
	}

	conn.SetMaxOpenConns(32)

	// Check if database running
	if err := waitForDb(conn.DB); err != nil {
		return nil, err
	}

	// Migrate database schema
	if err := migrateDb(conn.DB); err != nil {
		return nil, errors.Wrap(err, "could not migrate database.")
	}

	return conn, nil
}

// New creates a new database
func New() (Database, error) {
	conn, err := Connect()
	if err != nil {
		return nil, err
	}

	d := &database{
		conn: conn,
	}

	return d, nil
}

func waitForDb(conn *sql.DB) error {
	ready := make(chan struct{})
	go func() {
		for {
			if err := conn.Ping(); err == nil {
				close(ready)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-ready:
		return nil
	case <-time.After(time.Duration(*databaseTimeout) * time.Millisecond):
		return errors.New("Database not ready")
	}
}
