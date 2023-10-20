package main

import (
	"log"
	"os"
)

func main() {
	logger := log.New(os.Stdout, "[FIN] ", log.LstdFlags)
	defaultInterface := GetMaxInterface()
	if defaultInterface == "" {
		logger.Fatal("Could not get default interface")
	}

	logger.Printf("Using interface %v\n", defaultInterface)

	packetChan := make(chan PacketInfo, 65536)
	ipChan := make(chan map[string]uint64)
	countryChan := make(chan map[string]uint64)
	go Ingestion(packetChan, defaultInterface, "")
	go IpBuffer(packetChan, ipChan)
	go Geolocation(ipChan, countryChan)

	for currentMap := range countryChan {
		logger.Println(currentMap)
	}
}
