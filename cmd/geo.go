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

type geoInfo struct {
	Country        string
	BytesExchanged uint64
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

func Geolocation(in <-chan map[string]uint64, out chan<- map[string]uint64) {
	logger := log.New(os.Stdout, "[GEO] ", log.LstdFlags)
	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}

	logger.Println("Started Geolocation")
	for buffer := range in {
		logger.Printf("Received buffer %v\n", buffer)
		countriesData := make(map[string]uint64)
		for ip := range buffer {
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
				logger.Printf("Could not get country for `%v`: %v\n", ip, err)

			}

			if !data.Valid {
				logger.Printf("%v is not a valid public IP\n", ip)
				continue
			}

			_, exists = countriesData[data.Country]
			if exists {
				countriesData[data.Country] += bytesExchanged
			} else {
				countriesData[data.Country] = bytesExchanged
			}

			logger.Printf("%v | %v", ip, data.Country)
		}

		out <- countriesData
	}
}
