package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"
)

// User user profile
type User struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	BirthDate string `json:"birth_date"`
}

// Users map to user profile
type Users map[string]User

var gUsersMap = make(Users)
var gUsersMutex = &sync.Mutex{}
var gUsersFile = "users.json"

func main() {
	err := loadUsers()

	if err != nil {
		log.Fatal("Unable to load " + gUsersFile + ": " + err.Error())
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
	UUID := getUUID(r)

	if len(UUID) == 0 {
		http.Error(w, "Uuid not applied", http.StatusBadRequest)
		return
	}

	gUsersMutex.Lock()
	user, isPresent := gUsersMap[UUID]
	gUsersMutex.Unlock()

	if !isPresent {
		http.Error(w, "User not found", http.StatusNotFound)
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

	validUUID, _ := regexp.MatchString("[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}", user.UUID)
	if !validUUID {
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

	gUsersMutex.Lock()
	gUsersMap[user.UUID] = user
	saveUsers()
	gUsersMutex.Unlock()

	fmt.Fprintf(w, "OK")
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	UUID := getUUID(r)

	if len(UUID) == 0 {
		http.Error(w, "Uuid not applied", http.StatusBadRequest)
		return
	}

	gUsersMutex.Lock()
	delete(gUsersMap, UUID)
	saveUsers()
	gUsersMutex.Unlock()

	fmt.Fprintf(w, "OK")
}

func getUUID(r *http.Request) string {
	keys, ok := r.URL.Query()["uuid"]
	if ok && len(keys[0]) >= 1 {
		return keys[0]
	}
	return ""
}

func loadUsers() error {

	_, err := os.Stat(gUsersFile)
	if os.IsNotExist(err) {
		return nil
	}

	var body []byte
	body, err = ioutil.ReadFile(gUsersFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &gUsersMap)
	return err
}

func saveUsers() error {

	jsonString, err := json.Marshal(gUsersMap)
	if err != nil {
		return err
	}

	jsonData := []byte(jsonString)
	err = ioutil.WriteFile(gUsersFile, jsonData, 0644)

	return err
}
