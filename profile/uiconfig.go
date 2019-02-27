package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

var gUIConfigsMap = make(map[string][]byte)

func openUIConfig() error {
	var err error

	// 1. Check if /var/lib/profile exists
	configDirectory := uiConfigPathForFileName("")
	if _, err = os.Stat(configDirectory); os.IsNotExist(err) {
		return err
	}

	// 2. Ensure config files in /var/lib/profile
	var fileNames []string
	fileNames, err = filepath.Glob("uiconfig.*.json")
	if err != nil {
		return err
	}
	for _, fileName := range fileNames {
		configPath := uiConfigPathForFileName(fileName)
		if _, err = os.Stat(configPath); os.IsNotExist(err) {
			err = os.Link(fileName, configPath)
			if err != nil {
				return err
			}
		}
	}

	// 3. Load configs map
	var filePaths []string
	filePaths, err = filepath.Glob(uiConfigPathForFileName("uiconfig.*.json"))
	if err != nil {
		return err
	}
	for _, filePath := range filePaths {
		var body []byte
		body, err = ioutil.ReadFile(filePath)
		if err != nil {
			return err
		}
		_, fileName := filepath.Split(filePath)
		fileComponents := strings.Split(fileName, ".")
		lang := fileComponents[1]
		gUIConfigsMap[lang] = body
	}

	return nil
}

func uiConfigPathForFileName(fileName string) string {
	return filepath.Join(string(filepath.Separator), "var", "lib", "profile", fileName)
}

func uiConfigPathForLang(lang string) string {
	fileName := strings.Join([]string{"uiconfig", lang, "json"}, ".")
	return uiConfigPathForFileName(fileName)
}

func requestLang(r *http.Request) string {
	requestLang := "en"
	langs, isSet := r.URL.Query()["lang"]
	if isSet && (0 < len(langs)) && (0 < len(langs[0])) {
		lang := langs[0]
		if _, isSet = gUIConfigsMap[lang]; isSet {
			requestLang = lang
		}
	}
	return requestLang
}

func closeUIConfig() {
	gUIConfigsMap = nil
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
	configLang := requestLang(r)
	configData := gUIConfigsMap[configLang]
	if configData != nil {
		fmt.Fprintf(w, string(configData))
	} else {
		http.Error(w, "UI Config not set", http.StatusNotFound)
	}
}

func handleUIConfigPost(w http.ResponseWriter, r *http.Request) {

	configLang := requestLang(r)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	configPath := uiConfigPathForLang(configLang)
	err = ioutil.WriteFile(configPath, body, 0644)
	if err != nil {
		http.Error(w, "Failed to store ui config: "+err.Error(), http.StatusInternalServerError)
		return
	}

	gUIConfigsMap[configLang] = body
	fmt.Fprintf(w, "OK")
}

func handleUIConfigDelete(w http.ResponseWriter, r *http.Request) {

	configLang := requestLang(r)

	path := uiConfigPathForLang(configLang)

	err := os.Remove(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	delete(gUIConfigsMap, configLang)

	err = openUIConfig()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "OK")
}
