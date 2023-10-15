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

	ipChan := make(chan PacketInfo)
	go Ingestion(ipChan, defaultInterface, "")

	for ipStr := range ipChan {
		fmt.Println(ipStr)
	}
}
