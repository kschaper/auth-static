package services

import (
	"database/sql"
)

// CreateTableUsers is the SQL statement to create the users table.
const CreateTableUsers = `CREATE TABLE IF NOT EXISTS users (
	id  				TEXT NOT NULL PRIMARY KEY,
	email 			TEXT NOT NULL,
	code 				TEXT,
	hash				TEXT,
	created_at 	TEXT NOT NULL,
	updated_at 	TEXT,
	CONSTRAINT unique_email UNIQUE (email)
)`

// DatabaseClient creates tables and handles the database connection. It only supports SQLite.
type DatabaseClient struct {
	DSN string // data source name e.g. db filename or ":memory:"
}

// Open creates the database and tables and returns the database connection.
func (client *DatabaseClient) Open() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", client.DSN)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec(CreateTableUsers); err != nil {
		return nil, err
	}
	return db, nil
}
