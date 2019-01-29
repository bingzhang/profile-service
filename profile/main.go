package main

import (
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

func main() {

	var err error
	gDatabase, err = openDatabase(gDatabaseFile)
	if err != nil {
		log.Fatal("Failed to open " + gDatabaseFile + ": " + err.Error())
	}
	defer gDatabase.Close()

	http.HandleFunc("/profile", profileHandler)
	log.Fatal(http.ListenAndServe(":8082", nil))
}
