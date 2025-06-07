package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"encoding/hex"
	"errors"
	"net"
	"strings"
)

var macAddresses = map[string]string{
	"bmo":       "9c:6b:00:33:ef:18",
	"BMO":       "9c:6b:00:33:ef:18",
	"beemo":     "9c:6b:00:33:ef:18",
	"beemo-qc2": "9c:6b:00:33:ef:18",
}


// SendWOLPacket sends a Wake-on-LAN magic packet to the given MAC address
func SendWOLPacket(macAddr string) error {
	hwAddr, err := net.ParseMAC(macAddr)
	if err != nil {
		return err
	}
	if len(hwAddr) != 6 {
		return errors.New("invalid MAC address length")
	}
	// Magic packet: 6x 0xFF followed by MAC 16 times
	packet := make([]byte, 6+16*6)
	for i := 0; i < 6; i++ {
		packet[i] = 0xFF
	}
	for i := 0; i < 16; i++ {
		copy(packet[6+i*6:], hwAddr)
	}
	// UDP broadcast
	addr := &net.UDPAddr{
		IP:   net.IPv4bcast,
		Port: 9,
	}
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Write(packet)
	return err
}



func handleRequest(w http.ResponseWriter, r *http.Request) {
	alias := r.URL.Query().Get("alias")
	if alias == "" {
		http.Error(w, "Missing alias", http.StatusBadRequest)
		return
	}

	mac, exists := macAddresses[alias]
	if !exists {
		http.Error(w, "Unknown alias", http.StatusNotFound)
		return
	}

	err := SendWOLPacket(mac)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to send Wake-on-LAN packet: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Sent Wake-on-LAN packet to %s (%s)\n", alias, mac)
}

func main() {
	// Define a default port in case none is provided
	defaultPort := "8330"

	// Set up a flag to allow a port to be passed as a command-line argument
	port := flag.String("port", "", "Port for the WakeUp-LAN server to listen on")

	// Parse the command-line arguments
	flag.Parse()

	// If a port is provided as an argument, use it. If not, check the environment.
	if *port == "" {
		if envPort := os.Getenv("PORT"); envPort != "" {
			*port = envPort
		} else {
			*port = defaultPort
		}
	}

	// Output the port to indicate which one is being used
	fmt.Printf("Starting WakeUp-LAN Server on port %s!\n", *port)

	// Start the server on the specified port
	http.HandleFunc("/wakeonlan", handleRequest)
	err := http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", *port), nil)
	if err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
