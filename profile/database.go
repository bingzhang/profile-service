package main

import (
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// User Database
var gDatabase *sql.DB
var gDatabaseFile = "profile.db"

func openDatabase() error {

	path := filepath.Join(string(filepath.Separator), "var", "lib", "profile")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = gDatabaseFile
	} else {
		path = filepath.Join(path, gDatabaseFile)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return err
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE name='users'")
	if err != nil {
		return err
	}
	usersExists := rows.Next()
	rows.Close()

	if !usersExists {

		// Create New

		statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS users (
			uuid TEXT UNIQUE PRIMARY KEY NOT NULL DEFAULT (''),
			name TEXT NOT NULL DEFAULT (''),
			phone TEXT NOT NULL DEFAULT (''),
			birth_date TEXT NOT NULL DEFAULT (''),
			role INTEGER NOT NULL DEFAULT (0)
		)`)
		if err != nil {
			return err
		}

		_, err = statement.Exec()
		if err != nil {
			return err
		}

	} else {

		// Already Exits

		var roleExists int
		rows, err = db.Query("SELECT COUNT(*) AS COUNT FROM pragma_table_info('users') WHERE name='role'")
		if err == nil && rows.Next() {
			rows.Scan(&roleExists)
		}
		rows.Close()
		if roleExists == 0 {
			// Add 'role' column

			statement, err := db.Prepare("ALTER TABLE users ADD COLUMN role INTEGER NOT NULL DEFAULT (0)")
			if err != nil {
				return err
			}

			_, err = statement.Exec()
			if err != nil {
				return err
			}
		}
	}
	gDatabase = db
	return nil
}

func closeDatabase() {
	gDatabase.Close()
}
