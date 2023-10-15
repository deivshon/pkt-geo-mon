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

type GeoInfo struct {
	Country string
	Size    int
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

	return ipData{country, true}, nil
}

func Geolocation(in <-chan PacketInfo, out chan<- GeoInfo) {
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	httpClient := http.Client{
		Timeout: 2 * time.Second,
	}

	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	ipsCache := make(map[string]ipData)
	for {
		select {
		case <-ticker.C:
			for k := range ipsCache {
				delete(ipsCache, k)
			}
		case packet := <-in:
			logger.Printf("GOT %v\n", packet.DestinationIP)
			cachedIpData, exists := ipsCache[packet.DestinationIP]
			if exists && cachedIpData.Valid {
				logger.Printf("%v: %v (CACHED)\n", packet.DestinationIP, cachedIpData.Country)
				out <- GeoInfo{Country: cachedIpData.Country, Size: packet.PacketSize}
				continue
			} else if exists && !cachedIpData.Valid {
				logger.Printf("%v (CACHED, NOT VALID)\n", packet.DestinationIP)
				continue
			}

			ipData, err := getCountry(packet.DestinationIP, httpClient)
			if err != nil {
				logger.Printf("Could not get country for `%v`: %v\n", packet.DestinationIP, err)
				continue
			}

			ipsCache[packet.DestinationIP] = ipData
			if ipData.Valid {
				out <- GeoInfo{Country: ipData.Country, Size: packet.PacketSize}
				logger.Printf("%v: %v (API)\n", packet.DestinationIP, ipData.Country)
			} else {
				logger.Printf("%v (NOT VALID)\n", packet.DestinationIP)
			}
		}
	}
}
