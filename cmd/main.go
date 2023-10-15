package main

import (
	"fmt"
	"log"
)

func main() {
	defaultInterface := GetMaxInterface()
	if defaultInterface == "" {
		log.Fatal("Could not get default interface")
	}

	fmt.Printf("Using interface %v\n", defaultInterface)

	ipChan := make(chan PacketInfo, 65536)
	countryChan := make(chan GeoInfo)
	go Ingestion(ipChan, defaultInterface, "")
	go Geolocation(ipChan, countryChan)

	for country := range countryChan {
		if country.Size > 100000 {

		}
	}
}
