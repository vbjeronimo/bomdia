package main

import (
	"fmt"
	"net"
	"os"
)

func main() {
	args := os.Args
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s INTERFACE_NAME\n", args[0])
		os.Exit(1)
	}

	iface, err := net.InterfaceByName(args[1])
	if err != nil {
		panic(err)
	}

	groupAddr := net.UDPAddr{
		IP: net.ParseIP("244.0.0.251"),
		Port: 5353,
	}

	conn, err := net.ListenMulticastUDP("udp", iface, &groupAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	buf := make([]byte, iface.MTU)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		fmt.Printf("[%s:%d] %s (%d bytes)", src.IP, src.Port, buf[:n], n)
	}
}
