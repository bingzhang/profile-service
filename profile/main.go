package main

import (
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	var err error

	err = openDatabase()
	if err != nil {
		log.Fatal("Failed to open database: " + err.Error())
	} else {
		defer closeDatabase()
	}

	err = openUIConfig()
	if err != nil {
		log.Fatal("Failed to initialize UI config service: " + err.Error())
	} else {
		defer closeUIConfig()
	}

	http.HandleFunc("/profile", profileHandler)
	http.HandleFunc("/ui/config", uiConfigHandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
