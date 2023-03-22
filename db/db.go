package db

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/ethereum/go-ethereum/log"
)

// Postgres config
var (
	host     = "localhost"
	port     = 5432
	user     = "manav"
	dbname   = "postgres"
	password = ""
	sslmode  = "disable"
)

func SetupDB() (*sql.DB, error) {
	db, err := connect()
	if err != nil {
		log.Error("Unable to connect to db", "err", err)
		return nil, err
	}
	defer db.Close() // Use this in main loop

	// Ping the DB to make sure it's connected
	err = db.Ping()
	if err != nil {
		log.Error("Unable to ping DB", "err", err)
		return nil, err
	}

	log.Info("Successfully connect to postgres DB")

	// Create table for storing indexing runs
	err = createRunsTable(db)
	if err != nil {
		log.Error("Error in creating runs table", "err", err)
		return nil, err
	}

	// Create table for storing indxed data
	err = createIndexingTable(db)
	if err != nil {
		log.Error("Error in creating indexing table", "err", err)
		return nil, err
	}

	log.Info("Created required table for indexing service")

	return db, nil
}

func connect() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", host, port, user, dbname, sslmode)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}
	log.Info("Attempting to open postgres DB", "name", dbname)
	return sql.Open("postgres", connStr)
}

func createRunsTable(db *sql.DB) error {
	log.Debug("Attempting to create runs table")

	// Create the "runs" table if it does not exist
	query := `CREATE TABLE IF NOT EXISTS runs (
		id SERIAL PRIMARY KEY,
		start_block INTEGER NOT NULL,
		last_block INTEGER NOT NULL,
		contracts TEXT[]
	);`
	_, err := db.Exec(query)
	if err == nil {
		log.Debug("Runs table created successfully (if not already present)")
	}
	return err
}

func createIndexingTable(db *sql.DB) error {
	log.Debug("Attempting to create indexing table")

	// Create the "runs" table if it does not exist
	query := `CREATE TABLE IF NOT EXISTS indexooor (
		slot bytea,
		contract TEXT,
		value bytea,
		variable_name TEXT,
		key bytea,
		PRIMARY KEY (slot, contract)
	);`
	_, err := db.Exec(query)
	if err == nil {
		log.Debug("Indexing table created successfully (if not already present)")
	}
	return err
}
