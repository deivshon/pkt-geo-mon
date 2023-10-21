package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type GeoMap struct {
	CountryMap map[string]uint64
	Start      int64
	End        int64
}

type ipData struct {
	Country string
	Valid   bool
}

func getCountry(ip string, client http.Client) (ipData, error) {
	url := "http://ip-api.com/json/" + ip + "?fields=status,message,countryCode"

	response, err := client.Get(url)
	if err != nil {
		return ipData{"", true}, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return ipData{"", true}, err
	}

	var data map[string]string
	err = json.Unmarshal(body, &data)
	if err != nil {
		return ipData{"", true}, err
	}

	status, exists := data["status"]
	if !exists {
		return ipData{"", true}, fmt.Errorf("No `status` field in response")
	}
	if status != "success" {
		return ipData{"", false}, nil
	}

	country, exists := data["countryCode"]
	if !exists {
		return ipData{"", true}, fmt.Errorf("No `countryCode` field in response")
	}
	if country == "" {
		return ipData{"", false}, nil
	}

	return ipData{country, true}, nil
}

func logCountry(logger *log.Logger, twoLetterCode string, done int, total int, ip string) {
	if *verbose {
		logger.Printf("% 4d/%-4d | %v | %v", done, total, twoLetterCode, ip)
	}
}

func getTotals(buffer map[string]uint64) (int, uint64) {
	var valuesSum uint64 = 0
	for key := range buffer {
		valuesSum += buffer[key]
	}

	return len(buffer), valuesSum
}

func Geolocation(in <-chan IpMap, out chan<- GeoMap) {
	logger := log.New(os.Stdout, "[GEO] ", log.LstdFlags)
	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}

	logger.Println("Started Geolocation")
	for ipMap := range in {
		buffer := ipMap.Map
		ipsCount, totalBytes := getTotals(buffer)
		logger.Printf("Received map with %v ips and %v bytes exchanged", ipsCount, totalBytes)

		countriesData := make(map[string]uint64)
		doneCount := 0
		for ip := range buffer {
			time.Sleep(750 * time.Millisecond)
			bytesExchanged, exists := buffer[ip]
			if !exists {
				logger.Printf("Expected value for key %v\n", ip)
				continue
			}

			var data ipData
			for {
				currentTry, err := getCountry(ip, httpClient)
				if err == nil {
					data = currentTry
					break
				}
				logCountry(logger, "!!", doneCount, len(buffer), ip)
			}

			if !data.Valid {
				doneCount += 1
				logCountry(logger, "??", doneCount, len(buffer), ip)
				continue
			}

			_, exists = countriesData[data.Country]
			if exists {
				countriesData[data.Country] += bytesExchanged
			} else {
				countriesData[data.Country] = bytesExchanged
			}

			doneCount += 1
			logCountry(logger, data.Country, doneCount, len(buffer), ip)
		}

		out <- GeoMap{CountryMap: countriesData, Start: ipMap.Start, End: ipMap.End}
	}
}
