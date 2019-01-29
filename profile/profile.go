package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

func profileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "", "GET":
		handleProfileGet(w, r)
	case "POST":
		handleProfilePost(w, r)
	case "DELETE":
		handleProfileDelete(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusBadRequest)
	}

	//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handleProfileGet(w http.ResponseWriter, r *http.Request) {

	uuid := getUUID(r)

	if len(uuid) > 0 {
		handleProfileGetUser(w, uuid)
	} else {
		handleProfileGetUsers(w)
	}
}

func handleProfileGetUser(w http.ResponseWriter, uuid string) {

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

func handleProfileGetUsers(w http.ResponseWriter) {

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

func handleProfilePost(w http.ResponseWriter, r *http.Request) {

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

func handleProfileDelete(w http.ResponseWriter, r *http.Request) {
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
