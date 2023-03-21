package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
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

func connect() {
	connStr := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=%s", host, port, user, dbname, sslmode)
	if password != "" {
		connStr += fmt.Sprintf(" password=%s", password)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		log.Fatal(err)
	}

	// rows, err := db.Query("SELECT current_date;")
}
