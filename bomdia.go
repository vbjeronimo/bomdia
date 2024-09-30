package main

import (
	"encoding/binary"
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
		IP: net.ParseIP("224.0.0.251"),
		Port: 5353,
	}

	conn, err := net.ListenMulticastUDP("udp", iface, &groupAddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Printf("Listening for multicast DNS messages...\n")

	buf := make([]byte, iface.MTU)
	for {
		n, src, err := conn.ReadFromUDP(buf)
		if err != nil {
			panic(err)
		}

		if n < 12 {
			panic("Too few bytes received to be a proper mDNS packet")
		}

		ID := binary.BigEndian.Uint16(buf[0:2])

		QR := buf[2]&0x80 != 0
		OpCode := uint8(buf[2]>>3 & 0x0F)
		AA := buf[2]&0x04 != 0
		TC := buf[2]&0x02 != 0
		RD := buf[2]&0x01 != 0

		RA := buf[3]&0x80 != 0
		Z := uint8(buf[3]>>4 & 0x07)
		RCode := uint8(buf[3]&0x0F)

		NumQuestions := binary.BigEndian.Uint16(buf[4:6])
		NumAnswers := binary.BigEndian.Uint16(buf[6:8])
		NumAuthorityRR := binary.BigEndian.Uint16(buf[8:10])
		NumAdditionalRR := binary.BigEndian.Uint16(buf[10:12])

		fmt.Printf("Message from '%s:%d'\n", src.IP, src.Port)
		fmt.Printf("FULL HEADER: %b\n", buf[:12])

		fmt.Printf("Query ID:        %d\n", ID)

		fmt.Printf("QR Bit:          %v\n", QR)
		fmt.Printf("OpCode:          %d\n", OpCode)
		fmt.Printf("AA Bit:          %v\n", AA)
		fmt.Printf("TC Bit:          %v\n", TC)
		fmt.Printf("RD Bit:          %v\n", RD)

		fmt.Printf("RA Bit:          %v\n", RA)
		fmt.Printf("Z Bits:          %b\n", Z)
		fmt.Printf("RCode:           %d\n", RCode)

		fmt.Printf("# Questions:     %d\n", NumQuestions)
		fmt.Printf("# Answers:       %d\n", NumAnswers)
		fmt.Printf("# Authority RR:  %d\n", NumAuthorityRR)
		fmt.Printf("# Additional RR: %d\n", NumAdditionalRR)

		fmt.Printf("\n----------------\n")
		
	}
}
