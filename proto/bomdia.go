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

		QR := buf[2]>>7 & 0x01
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
		fmt.Printf("========== HEADER SECTION ==========\n")
		fmt.Printf("FULL HEADER: %b\n", buf[:12])

		fmt.Printf("Query ID:        %d\n", ID)

		fmt.Printf("QR Bit:          %b\n", QR)
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

		QName := ""
		offset := 12
		pos := offset
		for buf[pos] != '\x00' {
			length := int(buf[pos])
			pos++

			label := string(buf[pos : pos+length])
			QName += "." + label

			pos += length
		}
		QName = QName[1:]
		offset = pos + 1

		if QR == 0 {
			QType := binary.BigEndian.Uint16(buf[offset:offset+2])
			UnicastResponse := buf[offset+2] & 0x80 != 0
			QClassSection := buf[offset+2:offset+4]
			QClassSection[0] = QClassSection[0] & 0x7F
			QClass := binary.BigEndian.Uint16(QClassSection)

			fmt.Printf("========== QUESTION SECTION ==========\n")
			fmt.Printf("FULL SECTION:   %b\n", buf[12:n])
			fmt.Printf("QName:      %s\n", QName)
			fmt.Printf("QType:      %d\n", QType)
			fmt.Printf("Uni Res:    %v\n", UnicastResponse)
			fmt.Printf("QClass:     %d\n", QClass)
		} else {
			Type := binary.BigEndian.Uint16(buf[offset:offset+2])
			CacheFlush := buf[offset+2] & 0x80 != 0
			ClassSection := buf[offset+2:offset+4]
			ClassSection[0] = ClassSection[0] &0x7F
			Class := binary.BigEndian.Uint16(ClassSection)
			TTL := binary.BigEndian.Uint32(buf[offset+4:offset+8])
			RDLength := binary.BigEndian.Uint16(buf[offset+8:offset+10])

			RData := ""
			pos := offset + 10

			for _, b := range buf[pos:pos+int(RDLength)] {
				RData += fmt.Sprintf(".%d", b)
			}

			RData = RData[1:]

			fmt.Printf("========== ANSWER SECTION ==========\n")
			fmt.Printf("FULL SECTION:   %b\n", buf[12:n])
			fmt.Printf("Name:           %s\n", QName)
			fmt.Printf("Type:           %d\n", Type)
			fmt.Printf("Cache Flush:    %v\n", CacheFlush)
			fmt.Printf("Class:          %d\n", Class)
			fmt.Printf("TTL:            %d\n", TTL)
			fmt.Printf("RDLength:       %d\n", RDLength)
			fmt.Printf("RData:          %s\n", RData)
		}

		fmt.Printf("\n----------------\n")
		
	}
}
