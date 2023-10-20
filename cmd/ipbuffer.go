package main

import (
	"log"
	"os"
	"time"
)

func IpBuffer(in <-chan PacketInfo, out chan<- map[string]uint64) {
	logger := log.New(os.Stdout, "[IPB] ", log.LstdFlags)
	logger.Println("Started IpBuffer")
	ticker := time.NewTicker(12 * time.Hour)
	defer ticker.Stop()

	buffer := make(map[string]uint64)
	for {
		select {
		case <-ticker.C:
			logger.Printf("Flushing buffer %v\n", buffer)
			out <- buffer
			buffer = make(map[string]uint64)
		case packet := <-in:
			_, exists := buffer[packet.DestinationIP]
			if exists {
				buffer[packet.DestinationIP] += uint64(packet.PacketSize)
			} else {
				buffer[packet.DestinationIP] = uint64(packet.PacketSize)
			}
		}
	}
}
