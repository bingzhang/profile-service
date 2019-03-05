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

	// 1. Check if /var/lib/profile exists
	dbDirectory := filepath.Join(string(filepath.Separator), "var", "lib", "profile")
	if _, err := os.Stat(dbDirectory); os.IsNotExist(err) {
		return err
	}

	// 2. Check if /var/lib/profile/profile.db exists
	dbFile := filepath.Join(dbDirectory, gDatabaseFile)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		err = os.Link(dbFile, gDatabaseFile)
		if err != nil {
			return err
		}
	}

	// 3. Open database
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return err
	}

	// Check Users table
	err = checkUsers(db)
	if err != nil {
		return err
	}

	// Check Events table
	err = checkEvents(db)
	if err != nil {
		return err
	}

	gDatabase = db
	return nil
}

func closeDatabase() {
	gDatabase.Close()
}

func checkUsers(db *sql.DB) error {
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
	return nil
}

func checkEvents(db *sql.DB) error {
	rows, err := db.Query("SELECT name FROM sqlite_master WHERE name='events'")
	if err != nil {
		return err
	}
	eventsExists := rows.Next()
	rows.Close()

	if !eventsExists {

		// Create New

		statement, err := db.Prepare(`CREATE TABLE IF NOT EXISTS events (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL DEFAULT (''),
			time TEXT NOT NULL DEFAULT (''),
			duration INTEGER NOT NULL DEFAULT (0),
			
			location_description TEXT NOT NULL DEFAULT (''),
			location_latitude DOUBLE NOT NULL DEFAULT (0),
			location_longtitude DOUBLE NOT NULL DEFAULT (0),
			location_floor INTEGER NOT NULL DEFAULT (0),

			purchase_description TEXT NOT NULL DEFAULT (''),
			info_url TEXT NOT NULL DEFAULT (''),

			category TEXT NOT NULL DEFAULT (''),
			sub_category TEXT NOT NULL DEFAULT (''),

			user_role INTEGER NOT NULL DEFAULT (0)
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

	}
	return nil
}
