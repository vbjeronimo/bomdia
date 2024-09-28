package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "[ERROR] Could not get list of network interfaces: %s\n", err)
		os.Exit(1)
	}

	groupAddr := net.UDPAddr{
		IP: net.ParseIP("224.0.0.251"),
		Port: 5353,
	}

	for _, iface := range interfaces {
		if iface.Flags & net.FlagUp != 0 && iface.Flags & net.FlagRunning != 0 && iface.Flags & net.FlagLoopback == 0 {
			fmt.Printf("[INFO] Setting up interface '%s'\n", iface.Name)

			conn, err := net.ListenMulticastUDP("udp", &iface, &groupAddr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "[ERROR] Could not open an UDP connection for interface '%s': %s\n", iface.Name, err)
				os.Exit(1)
			}
			defer conn.Close()

			fmt.Printf("[SUCCESS] Network interface '%s' is listening on the mDNS multicast group\n", iface.Name)

			// TODO: extract this block into a function and run it in a goroutine
			buf := make([]byte, iface.MTU)
			for {
				n , src, err := conn.ReadFromUDP(buf)
				if err != nil {
					fmt.Fprintf(os.Stderr, "[ERROR] Failed to read packet: %s\n", err)
					os.Exit(1)
				}

				fmt.Printf("[%s:%d] %s (%d bytes)\n", src.IP, src.Port, buf[:n], n)
			}
		}
	}
}
