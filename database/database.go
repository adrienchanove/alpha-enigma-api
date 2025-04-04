package database

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(filepath string) {
	var err error
	DB, err = sql.Open("sqlite3", filepath)
	if err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	createUserTable := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE
	);
	`

	createMessageTable := `
	CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT,
		sender_id INTEGER,
		receiver_id INTEGER,
		FOREIGN KEY (sender_id) REFERENCES users(id),
		FOREIGN KEY (receiver_id) REFERENCES users(id)
	);
	`

	_, err := DB.Exec(createUserTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createMessageTable)
	if err != nil {
		log.Fatal(err)
	}
}
