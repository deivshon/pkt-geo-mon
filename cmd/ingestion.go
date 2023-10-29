package main

import (
	"log"
	"os"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
)

type PacketInfo struct {
	DestinationIP string
	PacketSize    int
}

func Ingestion(out chan<- PacketInfo, networkInterface string, bpfFilter string) {
	logger := log.New(os.Stdout, "[ING] ", log.LstdFlags)
	logger.Println("Started Ingestion")

	handle, err := pcap.OpenLive(networkInterface, 65575, true, pcap.BlockForever)
	if err != nil {
		logger.Fatalf("Error starting packets capture: %v", err)
	}
	defer handle.Close()

	if err := handle.SetBPFFilter(bpfFilter); err != nil {
		logger.Fatalf("Error setting BPF filter: %v", err)
	} else if bpfFilter != "" {
		logger.Printf("Using BPF filter %v", bpfFilter)
	}

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	for packet := range packetSource.Packets() {
		packetSize := len(packet.Data())
		if ip4Layer := packet.Layer(layers.LayerTypeIPv4); ip4Layer != nil {
			ip4, _ := ip4Layer.(*layers.IPv4)
			out <- PacketInfo{DestinationIP: ip4.DstIP.String(), PacketSize: packetSize}
		} else if ip6Layer := packet.Layer(layers.LayerTypeIPv6); ip6Layer != nil {
			ip6, _ := ip6Layer.(*layers.IPv6)
			out <- PacketInfo{DestinationIP: ip6.DstIP.String(), PacketSize: packetSize}
		}
	}
}
