package main

import (
	"net/http"
	"regexp"
)

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
