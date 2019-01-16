package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// User struct
type User struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
	Role      string `json:"role"`
}

// UserRole enum
type UserRole int

// UserRole values
const (
	UserRoleUnknown UserRole = iota
	UserRoleStudent
	UserRoleStaff
	UserRoleOther
)

// User Database
var gDatabase *sql.DB
var gDatabaseFile = "profile.db"

func main() {

	var err error
	gDatabase, err = openDatabase(gDatabaseFile)
	if err != nil {
		log.Fatal("Failed to open " + gDatabaseFile + ": " + err.Error())
	}
	defer gDatabase.Close()

	http.HandleFunc("/profile", handler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func openDatabase(fileName string) (*sql.DB, error) {

	path := filepath.Join(string(filepath.Separator), "var", "lib", "profile")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = fileName
	} else {
		path = filepath.Join(path, fileName)
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	rows, err := db.Query("SELECT name FROM sqlite_master WHERE name='users'")
	if err != nil {
		return nil, err
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
			return nil, err
		}

		_, err = statement.Exec()
		if err != nil {
			return nil, err
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
				return nil, err
			}

			_, err = statement.Exec()
			if err != nil {
				return nil, err
			}
		}
	}
	return db, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "", "GET":
		handleGet(w, r)
	case "POST":
		handlePost(w, r)
	case "DELETE":
		handleDelete(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusBadRequest)
	}

	//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handleGet(w http.ResponseWriter, r *http.Request) {

	uuid := getUUID(r)

	if len(uuid) > 0 {
		handleGetUser(w, uuid)
	} else {
		handleGetUsers(w)
	}
}

func handleGetUser(w http.ResponseWriter, uuid string) {

	if !isValidUUID(uuid) {
		http.Error(w, "Uuid not applied", http.StatusBadRequest)
		return
	}

	rows, err := gDatabase.Query("SELECT uuid, name, phone, birth_date, role FROM users WHERE uuid=\"" + uuid + "\"")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	if !rows.Next() {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user User
	var userRole UserRole
	err = rows.Scan(&user.UUID, &user.Name, &user.Phone, &user.BirthDate, &userRole)

	user.Role = userRoleToString(userRole)

	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data, err := json.Marshal(user)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, string(data))
}

func handleGetUsers(w http.ResponseWriter) {

	rows, err := gDatabase.Query("SELECT uuid, name, phone, birth_date, role FROM users")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "[")
	var count = 0
	for rows.Next() {
		var user User
		var userRole UserRole
		err = rows.Scan(&user.UUID, &user.Name, &user.Phone, &user.BirthDate, &userRole)

		user.Role = userRoleToString(userRole)

		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(user)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if count > 0 {
			fmt.Fprintf(w, ",")
		}

		fmt.Fprintf(w, string(data))
		count++
	}
	fmt.Fprintf(w, "]")
}

func handlePost(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if !isValidUUID(user.UUID) {
		http.Error(w, "Invalid uuid paramter", http.StatusBadRequest)
		return
	}

	validName, _ := regexp.MatchString("[a-zA-Z]+ [a-zA-Z]+", user.Name)
	if !validName {
		http.Error(w, "Invalid name paramter", http.StatusBadRequest)
		return
	}

	validPhone, _ := regexp.MatchString("^(\\+?1\\s?)?((\\([0-9]{3}\\))|[0-9]{3})[\\s\\-]?[\\0-9]{3}[\\s\\-]?[0-9]{4}$", user.Phone)
	if !validPhone {
		http.Error(w, "Invalid phone paramter", http.StatusBadRequest)
		return
	}

	_, err = time.Parse("2006/01/02", user.BirthDate)
	if err != nil {
		http.Error(w, "Invalid birth_date paramter", http.StatusBadRequest)
		return
	}

	userRole := userRoleFromString(user.Role)
	if userRole == UserRoleUnknown {
		http.Error(w, "Invalid role paramter", http.StatusBadRequest)
		return
	}

	rows, err := gDatabase.Query("SELECT uuid FROM users WHERE uuid=\"" + user.UUID + "\"")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userExists := rows.Next()
	rows.Close()

	if userExists {

		// Update Existing

		statement, err := gDatabase.Prepare("UPDATE users SET name=?, phone=?, birth_date=?, role=? WHERE uuid=\"" + user.UUID + "\"")
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(user.Name, user.Phone, user.BirthDate, userRole)
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {

		// Create New

		statement, err := gDatabase.Prepare("INSERT INTO users(uuid, name, phone, birth_date, role) values(?,?,?,?,?)")
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(user.UUID, user.Name, user.Phone, user.BirthDate, userRole)
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "OK")
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	uuid := getUUID(r)

	if !isValidUUID(uuid) {
		http.Error(w, "Uuid not applied", http.StatusBadRequest)
		return
	}

	statement, err := gDatabase.Prepare("DELETE FROM users WHERE uuid=\"" + uuid + "\"")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	_, err = statement.Exec()
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OK")
}

func getUUID(r *http.Request) string {
	keys, ok := r.URL.Query()["uuid"]
	if ok && len(keys[0]) >= 1 {
		return keys[0]
	}
	return ""
}

func isValidUUID(uuid string) bool {
	isValid, _ := regexp.MatchString("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}", uuid)
	return isValid
}

func userRoleFromString(value string) UserRole {
	if value == "student" {
		return UserRoleStudent
	} else if value == "staff" {
		return UserRoleStaff
	} else if value == "other" {
		return UserRoleOther
	} else {
		return UserRoleUnknown
	}
}

func userRoleToString(value UserRole) string {
	switch value {
	case UserRoleStudent:
		return "student"
	case UserRoleStaff:
		return "staff"
	case UserRoleOther:
		return "other"
	default:
		return ""
	}
}
