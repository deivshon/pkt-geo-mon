package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

func handleRequest(logger *log.Logger, db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		startInt64, err := strconv.ParseInt(start, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid start parameter: %v", err), http.StatusBadRequest)
			return
		}
		endInt64, err := strconv.ParseInt(end, 10, 64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Invalid end parameter: %v", err), http.StatusBadRequest)
			return
		}

		countriesMap, err := GetCountrySum(logger, db, startInt64, endInt64)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not get countries map: %v", err), http.StatusInternalServerError)
			return
		}

		countriesMapJson, err := json.Marshal(countriesMap)
		if err != nil {
			http.Error(w, fmt.Sprintf("Could not marshal countries map result: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(countriesMapJson)
	}

}

func Api(db *sql.DB) {
	logger := log.New(os.Stdout, "[API] ", log.LstdFlags)

	logger.Println("Started API")
	http.HandleFunc("/data", handleRequest(logger, db))
	http.ListenAndServe(":8080", nil)
}
