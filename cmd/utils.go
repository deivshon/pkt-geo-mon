package main

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/gopacket/pcap"
)

func readSysFile(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func getInterfaceBytes(interfaceName string) (int64, error) {
	rxPath := filepath.Join("/sys/class/net/", interfaceName, "/statistics/rx_bytes")
	txPath := filepath.Join("/sys/class/net/", interfaceName, "/statistics/tx_bytes")

	rxBytes, err := readSysFile(rxPath)
	if err != nil {
		return 0, err
	}

	txBytes, err := readSysFile(txPath)
	if err != nil {
		return 0, err
	}

	rxBytes = strings.TrimSpace(rxBytes)
	txBytes = strings.TrimSpace(txBytes)

	rxBytesInt, err := strconv.ParseInt(rxBytes, 10, 64)
	if err != nil {
		return 0, err
	}

	txBytesInt, err := strconv.ParseInt(txBytes, 10, 64)
	if err != nil {
		return 0, err
	}

	return rxBytesInt + txBytesInt, nil
}

func GetMaxInterface() string {
	interfaces, err := pcap.FindAllDevs()
	if err != nil {
		log.Fatal(err)
	}

	if len(interfaces) == 0 {
		log.Fatal("No network interfaces found")
	}

	maxInterface := ""
	var maxInterfaceBytes int64 = 0
	for _, ifa := range interfaces {
		interfaceBytes, err := getInterfaceBytes(ifa.Name)
		if err != nil {
			continue
		}

		if interfaceBytes > maxInterfaceBytes {
			maxInterfaceBytes = interfaceBytes
			maxInterface = ifa.Name
		}

	}

	return maxInterface
}
