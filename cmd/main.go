package main

import (
	"flag"
	"log"
	"os"
)

var verbose = flag.Bool("v", false, "Verbose output")
var bufferPeriod = flag.Int("p", 600, "Frequency of IPs buffer flushing in seconds")
var bpfFilter = flag.String("f", "", "BPF expression for filtering packet captures")

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "[FIN] ", log.LstdFlags)
	defaultInterface := GetMaxInterface()
	if defaultInterface == "" {
		logger.Fatal("Could not get default interface")
	}

	logger.Printf("Using interface %v\n", defaultInterface)
	logger.Printf("Using period of %v seconds\n", *bufferPeriod)
	if *verbose {
		logger.Printf("Using verbose output")
	}
	if *bpfFilter != "" {
		logger.Printf("Using BPF filter `%v`", *bpfFilter)
	}
	packetChan := make(chan PacketInfo, 65536)
	ipChan := make(chan map[string]uint64)
	countryChan := make(chan map[string]uint64)
	go Ingestion(packetChan, defaultInterface, *bpfFilter)
	go IpBuffer(packetChan, ipChan, *bufferPeriod)
	go Geolocation(ipChan, countryChan)

	for currentMap := range countryChan {
		logger.Println(currentMap)
	}
}
