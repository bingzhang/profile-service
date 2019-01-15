package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// User profile
type User struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
}

// User Database
var gDatabase *sql.DB
var gDatabaseFile = "profile.db"

func main() {
	var err error

	gDatabase, err = sql.Open("sqlite3", gDatabaseFile)
	if err != nil {
		log.Fatal("Failed to open " + gDatabaseFile + ": " + err.Error())
	}

	statement, err := gDatabase.Prepare("CREATE TABLE IF NOT EXISTS users (uuid TEXT PRIMARY KEY, name TEXT, phone TEXT, birth_date TEXT)")
	if err != nil {
		log.Fatal("Unable to access " + gDatabaseFile + ": " + err.Error())
	}

	_, err = statement.Exec()
	if err != nil {
		log.Fatal("Unable to access " + gDatabaseFile + ": " + err.Error())
	}

	http.HandleFunc("/profile", handler)
	log.Fatal(http.ListenAndServe(":8082", nil))
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

	if !isValidUUID(uuid) {
		http.Error(w, "Uuid not applied", http.StatusBadRequest)
		return
	}

	rows, err := gDatabase.Query("SELECT uuid, name, phone, birth_date FROM users WHERE uuid=\"" + uuid + "\"")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if !rows.Next() {
		rows.Close()
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	var user User
	err = rows.Scan(&user.UUID, &user.Name, &user.Phone, &user.BirthDate)
	rows.Close()

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

	rows, err := gDatabase.Query("SELECT uuid FROM users WHERE uuid=\"" + user.UUID + "\"")
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}

	userExists := rows.Next()
	rows.Close()

	if userExists {

		// Update Existing

		statement, err := gDatabase.Prepare("UPDATE users SET name=?, phone=?, birth_date=? WHERE uuid=\"" + user.UUID + "\"")
		if err != nil {
			http.Error(w, "Internal Error Occured1: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(user.Name, user.Phone, user.BirthDate)
		if err != nil {
			http.Error(w, "Internal Error Occured2: "+err.Error(), http.StatusInternalServerError)
			return
		}
	} else {

		// Create New

		statement, err := gDatabase.Prepare("INSERT INTO users(uuid, name, phone, birth_date) values(?,?,?,?)")
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(user.UUID, user.Name, user.Phone, user.BirthDate)
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
