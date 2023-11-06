package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
)

func initStorage(logger *log.Logger, db *sql.DB) {
	sqlStmt := `
	CREATE TABLE IF NOT EXISTS measurements (
		MeasurementID INTEGER PRIMARY KEY,
		StartTime INTEGER NOT NULL,
		EndTime INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS bytesExchanged (
		MeasurementID INTEGER NOT NULL,
		CountryCode TEXT NOT NULL,
		BytesExchanged INTEGER NOT NULL,
		FOREIGN KEY(MeasurementID) REFERENCES measurements(MeasurementID)
	);
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		logger.Fatalf("Could not initialize DB: %v", err)
	}
}

func createMeasurement(db *sql.DB, start int64, end int64) (int64, error) {
	creationStmt := `INSERT INTO measurements (StartTime, EndTime) VALUES (?, ?)`
	_, err := db.Exec(creationStmt, start, end)
	if err != nil {
		return 0, err
	}

	var id int64
	idStmt := `SELECT last_insert_rowid()`
	err = db.QueryRow(idStmt).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func Storage(db *sql.DB, in <-chan GeoMap) {
	logger := log.New(os.Stdout, "[STG] ", log.LstdFlags)
	logger.Println("Started Storage")

	initStorage(logger, db)

	insertStmt := `INSERT INTO bytesExchanged (MeasurementID, CountryCode, BytesExchanged) VALUES (?, ?, ?)`
	errorOccurred := false
	for geoData := range in {
		logger.Println("Starting save")
		id, err := createMeasurement(db, geoData.Start, geoData.End)
		if err != nil {
			logger.Fatalf("Error occurred while creating measurement data: %v", err)
		}

		for countryCode, bytesExchanged := range geoData.CountryMap {
			_, err := db.Exec(insertStmt, id, countryCode, bytesExchanged)
			if err != nil {
				errorOccurred = false
				logger.Printf("Error occurred saving country measurement: %v", err)
			}
		}

		if errorOccurred {
			logger.Fatal("Shutting down due to storage error")
		}
		logger.Printf("Ended save for data with id %v", id)
	}
}

func GetCountrySum(logger *log.Logger, db *sql.DB, start int64, end int64) (map[string]uint64, error) {
	query := `SELECT b.CountryCode, SUM(b.BytesExchanged) as TotalBytesExchanged
	FROM BytesExchanged b
	JOIN Measurements m ON b.MeasurementID = m.MeasurementID
	WHERE m.StartTime >= ?
	AND m.EndTime <= ?
	GROUP BY b.CountryCode;
	`

	rows, err := db.Query(query, start, end)
	if err != nil {
		return nil, fmt.Errorf("could not query DB: %v", err)
	}
	defer rows.Close()

	countriesMap := make(map[string]uint64)
	for rows.Next() {
		var countryCode string
		var totalBytesExchanged uint64
		err := rows.Scan(&countryCode, &totalBytesExchanged)
		if err != nil {
			return nil, fmt.Errorf("could not scan row: %v", err)
		}
		countriesMap[countryCode] = totalBytesExchanged
	}

	return countriesMap, nil
}
