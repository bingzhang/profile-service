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

	http.HandleFunc("/profile", profileHandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
