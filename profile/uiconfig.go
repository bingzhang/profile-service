package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

var gUIConfig []byte
var gUIConfigFile = "uiconfig.json"

func openUIConfig() error {
	path := uiConfigPath()

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		err = os.Link(gUIConfigFile, path)
		if err != nil {
			return err
		}
	}

	var body []byte
	body, err = ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	gUIConfig = body
	return nil
}

func saveUIConfig() error {
	return saveUIConfigData(gUIConfig)
}

func saveUIConfigData(data []byte) error {
	return ioutil.WriteFile(uiConfigPath(), data, 0644)
}

func closeUIConfig() {
	gUIConfig = nil
}

func uiConfigPath() string {
	path := filepath.Join(string(filepath.Separator), "var", "lib", "profile")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = gUIConfigFile
	} else {
		path = filepath.Join(path, gUIConfigFile)
	}
	return path
}

func uiConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "", "GET":
		handleUIConfigGet(w, r)
	case "POST":
		handleUIConfigPost(w, r)
	case "DELETE":
		handleUIConfigDelete(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusBadRequest)
	}

	//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handleUIConfigGet(w http.ResponseWriter, r *http.Request) {
	if gUIConfig != nil {
		fmt.Fprintf(w, string(gUIConfig))
	} else {
		http.Error(w, "UI Config not set", http.StatusNotFound)
	}
}

func handleUIConfigPost(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = saveUIConfigData(body)
	if err != nil {
		http.Error(w, "Failed to save ui config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gUIConfig = body
	fmt.Fprintf(w, "OK")
}

func handleUIConfigDelete(w http.ResponseWriter, r *http.Request) {
	path := uiConfigPath()

	if path == gUIConfigFile {
		http.Error(w, "Unable to perform operation", http.StatusBadRequest)
		return
	}

	err := os.Remove(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = openUIConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OK")
}
