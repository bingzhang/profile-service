package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	_ "github.com/mattn/go-sqlite3"
)

// Location struct
type Location struct {
	Desription string  `json:"description"`
	Latitude   float64 `json:"latitude"`
	Longtitude float64 `json:"longtitude"`
	Floor      int32   `json:"floor"`
}

// Event struct
type Event struct {
	ID                  int32    `json:"id"`
	Name                string   `json:"name"`
	Time                string   `json:"time"`
	Duration            int32    `json:"duration"`
	Location            Location `json:"location"`
	PurchaseDescription string   `json:"purchase_description"`
	InfoURL             string   `json:"info_url"`
	Category            string   `json:"category"`
	SubCategory         string   `json:"sub_category"`
	UserRole            string   `json:"user_role"`
}

func eventsHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "", "GET":
		handleEventsGet(w, r)
	case "POST":
		handleEventsPost(w, r)
	case "DELETE":
		handleEventsDelete(w, r)
	default:
		http.Error(w, "Method not supported", http.StatusBadRequest)
	}

	//	fmt.Fprintf(w, "Hi there, I love %s!", r.URL.Path[1:])
}

func handleEventsGet(w http.ResponseWriter, r *http.Request) {

	where := ""

	// SELECT * from events WHERE (time < datetime("2019-03-05 15:00:00", "start of day", "1 day")) AND ("2019-03-05 15:00:00" < datetime(time, duration || " minutes"))

	time := getTime(r)
	if isValidTime(time) {
		timeClause := fmt.Sprintf("(time < datetime(\"%s\", \"start of day\", \"1 day\")) AND (\"%s\" < datetime(time, duration || \" minutes\"))", time, time)
		if 0 < len(where) {
			where += " AND"
		}
		where += timeClause
	}

	userRole := userRoleFromString(getUserRole(r))
	if userRole != UserRoleUnknown {
		userRoleClause := fmt.Sprintf("(user_role = %d)", userRole)
		if 0 < len(where) {
			where += " AND"
		}
		where += userRoleClause
	}

	query := "SELECT id, name, time, duration, location_description, location_latitude, location_longtitude, location_floor, purchase_description, info_url, category, sub_category, user_role FROM events "
	if 0 < len(where) {
		query += "WHERE " + where
	}

	rows, err := gDatabase.Query(query)
	if err != nil {
		http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	fmt.Fprintf(w, "[")
	var count = 0
	for rows.Next() {
		var event Event
		var userRole UserRole
		err = rows.Scan(&event.ID, &event.Name, &event.Time, &event.Duration,
			&event.Location.Desription, &event.Location.Latitude, &event.Location.Longtitude, &event.Location.Floor,
			&event.PurchaseDescription, &event.InfoURL, &event.Category, &event.SubCategory,
			&userRole)

		event.UserRole = userRoleToString(userRole)

		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		data, err := json.Marshal(event)

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

func handleEventsPost(w http.ResponseWriter, r *http.Request) {

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var events []Event
	err = json.Unmarshal(body, &events)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if 0 < len(events) {

		sql := "INSERT INTO events(name, time, duration, location_description, location_latitude, location_longtitude, location_floor, purchase_description, info_url, category, sub_category, user_role) VALUES "
		values := []interface{}{}

		for _, event := range events {
			userRole := userRoleFromString(event.UserRole)
			sql += "(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?),"

			values = append(values, event.Name, event.Time, event.Duration,
				event.Location.Desription, event.Location.Latitude, event.Location.Longtitude, event.Location.Floor,
				event.PurchaseDescription, event.InfoURL,
				event.Category, event.SubCategory,
				userRole)
		}
		sql = sql[0 : len(sql)-1] // trim the last ,

		statement, err := gDatabase.Prepare(sql)
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}

		_, err = statement.Exec(values...)
		if err != nil {
			http.Error(w, "Internal Error Occured: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	fmt.Fprintf(w, "OK")
}

func handleEventsDelete(w http.ResponseWriter, r *http.Request) {
	ids := getIds(r)
	sql := "DELETE FROM events"
	if 0 < len(ids) {
		sql = "DELETE FROM events WHERE id IN (" + ids + ")"
	}
	statement, err := gDatabase.Prepare(sql)
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

// Utils

func getTime(r *http.Request) string {
	times, isSet := r.URL.Query()["time"]
	if isSet && (0 < len(times)) {
		return times[0]
	}
	return ""
}

func isValidTime(time string) bool {
	// 2019-03-05 15:30:00
	isValid, _ := regexp.MatchString("[0-9]{4}-[0-9]{2}-[0-9]{2} [0-9]{2}:[0-9]{2}:[0-9]{2}", time)
	return isValid
}

func getUserRole(r *http.Request) string {
	userRoles, isSet := r.URL.Query()["role"]
	if isSet && (0 < len(userRoles)) {
		return userRoles[0]
	}
	return ""
}

func getIds(r *http.Request) string {
	ids, isSet := r.URL.Query()["id"]
	if isSet && (0 < len(ids)) {
		return ids[0]
	}
	return ""
}
