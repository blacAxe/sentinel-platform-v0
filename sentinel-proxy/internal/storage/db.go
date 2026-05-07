package db

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func Init() {
	var err error
	DB, err = sql.Open("sqlite3", "data/sentinel.db")
	if err != nil {
		log.Fatal(err)
	}

	createTable()
}

func createTable() {
	query := `
	CREATE TABLE IF NOT EXISTS logs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp TEXT,
		ip TEXT,
		path TEXT,
		query TEXT,
		action TEXT,
		reason TEXT
	);`

	_, err := DB.Exec(query)
	if err != nil {
		log.Fatal(err)
	}
}