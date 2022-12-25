package database

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/namsral/flag"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

var (
	databaseURL     = flag.String("database-url", "postgres://postgres:asgu@2409@localhost:5432/postgres?sslmode=disable", "Database URL.")
	databaseTimeout = flag.Int64("database-timeout-ms", 2000, "")
)

// Connect creates a new database connection
func Connect() (*sqlx.DB, error) {
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
