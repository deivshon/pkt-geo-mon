package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var verbose = flag.Bool("v", false, "Verbose output")
var bufferPeriod = flag.Int("p", 600, "Frequency of IPs buffer flushing in seconds")
var bpfFilter = flag.String("f", "", "BPF expression for filtering packet captures")
var dbPath = flag.String("d", "bytesgeo.db", "Path to the sqlite db file")

func main() {
	flag.Parse()

	logger := log.New(os.Stdout, "[FIN] ", log.LstdFlags)
	defaultInterface := GetMaxInterface()
	if defaultInterface == "" {
		logger.Fatal("Could not get default interface")
	}

	db, err := sql.Open("sqlite3", *dbPath)
	if err != nil {
		logger.Fatalf("Could not open DB: %v", err)
	}
	defer db.Close()

	logger.Printf("Using interface %v", defaultInterface)
	logger.Printf("Using DB %v", *dbPath)
	logger.Printf("Using period of %v seconds", *bufferPeriod)
	if *verbose {
		logger.Printf("Using verbose output")
	}
	if *bpfFilter != "" {
		logger.Printf("Using BPF filter %v", *bpfFilter)
	}

	packetChan := make(chan PacketInfo, 65536)
	ipChan := make(chan IpMap)
	countryChan := make(chan GeoMap)
	go Ingestion(packetChan, defaultInterface, *bpfFilter)
	go IpBuffer(packetChan, ipChan, *bufferPeriod)
	go Geolocation(ipChan, countryChan)
	go Storage(db, countryChan)
	go Api(db)

	select {}
}
