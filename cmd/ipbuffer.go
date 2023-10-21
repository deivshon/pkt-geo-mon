package main

import (
	"log"
	"os"
	"time"
)

type IpMap struct {
	Map   map[string]uint64
	Start int64
	End   int64
}

func IpBuffer(in <-chan PacketInfo, out chan<- IpMap, period int) {
	logger := log.New(os.Stdout, "[IPB] ", log.LstdFlags)
	logger.Println("Started IpBuffer")
	ticker := time.NewTicker(time.Duration(period) * time.Second)
	defer ticker.Stop()

	buffer := make(map[string]uint64)
	start := time.Now().Unix()
	for {
		select {
		case <-ticker.C:
			nowUnix := time.Now().Unix()

			logger.Println("Flushing buffer...")
			out <- IpMap{Map: buffer, Start: start, End: nowUnix}
			logger.Println("Buffer flushed")

			start = nowUnix
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
