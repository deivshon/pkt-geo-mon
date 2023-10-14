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

	ipChan := make(chan string)
	go Ingestion(ipChan, defaultInterface, "")

	for ipStr := range ipChan {
		fmt.Println(ipStr)
	}
}
