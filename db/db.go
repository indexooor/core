package db

import (
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	log "github.com/inconshreveable/log15"
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

type DB struct {
	db              *sql.DB
	runInsert       *sql.Stmt
	runUpdate       *sql.Stmt
	indexooorInsert *sql.Stmt
}

type Run struct {
	Id         uint64         `db:"id"`
	StartBlock uint64         `db:"start_block"`
	LastBlock  uint64         `db:"end_block"`
	Contracts  pq.StringArray `db:"contracts"`
}
type Indexooor struct {
	Slot         string `db:"slot"`
	Contract     string `db:"contract"`
	Value        string `db:"value"`
	VariableName string `db:"variable_name"`
	Key          string `db:"key"`
}

func SetupDB() (*DB, error) {
	db, err := connect()
	if err != nil {
		log.Error("Unable to connect to db", "err", err)
		return nil, err
	}

	// Ping the DB to make sure it's connected
	err = db.Ping()
	if err != nil {
		log.Error("Unable to ping DB", "err", err)
		return nil, err
	}

	log.Info("Successfully connect to postgres DB")

	obj := &DB{db: db}

	// Create table for storing indexing runs
	err = obj.createRunsTable()
	if err != nil {
		return nil, err
	}

	// Create table for storing indxed data
	err = obj.createIndexingTable()
	if err != nil {
		return nil, err
	}

	log.Info("Created required table for indexing service")

	err = obj.prepareStatements()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

func (db *DB) Close() {
	db.runInsert.Close()
	db.runUpdate.Close()
	db.indexooorInsert.Close()
	db.db.Close()
}

func (db *DB) prepareStatements() error {
	query := "INSERT INTO runs (start_block, last_block, contracts) VALUES ($1, $2, $3)"
	runInsertStatement, err := db.db.Prepare(query)
	if err != nil {
		log.Error("Error in creating insert statement for runs table", "err", err)
		return err
	}

	db.runInsert = runInsertStatement

	query = "UPDATE runs SET last_block=$1 WHERE id=$2"
	runUpdateStatement, err := db.db.Prepare(query)
	if err != nil {
		log.Error("Error in creating update statement for runs table", "err", err)
		return err
	}

	db.runUpdate = runUpdateStatement

	query = "INSERT INTO indexooor (slot, contract, value, variable_name, key) VALUES ($1, $2, $3, $4, $5)"
	indexooorInsertStatement, err := db.db.Prepare(query)
	if err != nil {
		log.Error("Error in creating insert statement for indexooor table", "err", err)
		return err
	}

	db.indexooorInsert = indexooorInsertStatement

	return nil
}

func connect() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", host, port, user, dbname, sslmode)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}
	log.Info("Attempting to open postgres DB", "name", dbname)
	return sql.Open("postgres", connStr)
}

func (db *DB) createRunsTable() error {
	log.Debug("Attempting to create runs table")

	// Create the "runs" table if it does not exist
	query := `CREATE TABLE IF NOT EXISTS runs (
		id SERIAL PRIMARY KEY,
		start_block INTEGER NOT NULL,
		last_block INTEGER NOT NULL,
		contracts TEXT[]
	);`

	_, err := db.db.Exec(query)
	if err != nil {
		log.Error("Error in creating runs table", "err", err)
		return err
	}

	log.Info("Runs table created successfully (if not already present)")

	return nil
}

func (db *DB) createIndexingTable() error {
	log.Debug("Attempting to create indexing table")

	// Create the "runs" table if it does not exist
	query := `CREATE TABLE IF NOT EXISTS indexooor (
		slot TEXT,
		contract TEXT,
		value TEXT,
		variable_name TEXT,
		key TEXT,
		PRIMARY KEY (slot, contract)
	);`

	_, err := db.db.Exec(query)
	if err != nil {
		log.Error("Error in creating indexing table", "err", err)
		return err
	}

	log.Info("Indexing table created successfully (if not already present)")

	return nil
}

func (db *DB) CreateNewRun(run *Run) error {
	log.Info("Attempting to create new run", "start block", run.StartBlock, "last block", run.LastBlock, "contracts", run.Contracts)

	_, err := db.runInsert.Exec(run.StartBlock, run.LastBlock, pq.Array(run.Contracts))
	if err != nil {
		log.Info("Error in creating new run", "err", err)
		return err
	}

	log.Info("Created new run entry")
	return nil
}

// UpdateRun updates an existing run instance given ID and last block (indexed) number
func (db *DB) UpdateRun(id uint64, lastBlock uint64) error {
	log.Debug("Attempting to update new run", "id", id, "last indexed block", lastBlock)

	res, err := db.runUpdate.Exec(lastBlock, id)
	if err != nil {
		log.Error("Error in updating existing run", "id", id, "err", err)
		return err
	}

	// Check if the query updated any rows
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		log.Error("Error in fetching rows affected", "id", id, "err", err)
		return err
	}

	if rowsAffected == 0 {
		log.Error("Run ID not found", "id", id)
		return fmt.Errorf("Run with ID: %d not found", id)
	}

	log.Debug("Updated existing run entry", "id", id)

	return nil
}

// FetchRunByID fetches an existing run from runs table by ID
func (db *DB) FetchRunByID(id uint64) (*Run, error) {
	var run Run
	err := db.db.QueryRow("SELECT id, start_block, last_block, contracts FROM runs WHERE id = $1", id).Scan(&run.Id, &run.StartBlock, &run.LastBlock, &run.Contracts)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Debug("No run with given ID found in runs table", "id", id)
		} else {
			log.Error("Unable to fetch run with given ID from runs table", "id", id, "err", err)
		}
	}

	return &run, err
}

func (db *DB) AddNewIndexingEntry(obj *Indexooor) error {
	log.Debug("Attempting to add new indexing row")

	var variableName interface{}
	if obj.VariableName != "" {
		variableName = obj.VariableName
	} else {
		variableName = nil
	}

	var key interface{}
	if obj.Key != "" {
		key = obj.Key
	} else {
		key = nil
	}

	_, err := db.indexooorInsert.Exec(obj.Slot, obj.Contract, obj.Value, variableName, key)
	if err != nil {
		log.Info("Error in creating new run", "err", err)
		return err
	}

	log.Info("Created new run entry")
	return nil
}
