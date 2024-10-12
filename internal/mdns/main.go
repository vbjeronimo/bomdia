package mdns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

var (
	ErrTooFewBytes = errors.New("too few bytes received to be a proper mDNS packet")
)

type MDNSMessage struct {
	Source net.UDPAddr

	ID              uint16
	Flags           MDNSFlags
	NumQuestions    uint16
	NumAnswers      uint16
	NumAuthorityRR  uint16
	NumAdditionalRR uint16

	Questions []MDNSQuestion
	Answers   []MDNSAnswer

	// Not sure yet if I'll need the other two slices, so I'll skip them for now
}

type MDNSFlags struct {
	Query               bool
	AuthoritativeAnswer bool
	Truncated           bool
	RecursionDesired    bool
	RecursionAvailable  bool

	OPCode       uint8
	ResponseCode uint8

	zero uint8 // The "Zero bits" are a reserved section that's not used for anything, so I won't be exporting it
}

type MDNSQuestion struct {
	Name  string
	Type  uint16
	Class uint16
}

type MDNSAnswer struct {
	Name     string
	Type     uint16
	Class    uint16
	TTL      uint32
	RDLength uint16
	RData    string
}

func DecodePacket(packet []byte, source net.UDPAddr) (*MDNSMessage, error) {
	if len(packet) < 12 {
		return nil, fmt.Errorf("%w: %d bytes received", ErrTooFewBytes, len(packet))
	}

	byteParser, err := newByteParser(packet)
	if err != nil {
		return nil, err
	}

	message := &MDNSMessage{Source: source}

	parseHeader(byteParser, message)

	if message.NumQuestions > 0 {
		parseQuestions(byteParser, message)
	}

	if message.NumAnswers > 0 {
		parseAnswers(byteParser, message)
	}

	return message, nil
}

func parseHeader(byteParser *byteParser, message *MDNSMessage) {
	message.ID = binary.BigEndian.Uint16(byteParser.readBytes(2))

	flagsSection := byteParser.readBytes(2)

	message.Flags.Query = flagsSection[0]&0x80 != 0
	message.Flags.OPCode = flagsSection[0] >> 3 & 0x0F
	message.Flags.AuthoritativeAnswer = flagsSection[0]&0x04 != 0
	message.Flags.Truncated = flagsSection[0]&0x02 != 0
	message.Flags.RecursionDesired = flagsSection[0]&0x01 != 0

	message.Flags.RecursionAvailable = flagsSection[1]&0x80 != 0
	message.Flags.zero = flagsSection[1] >> 4 & 0x07
	message.Flags.ResponseCode = flagsSection[1] & 0x0F

	message.NumQuestions = binary.BigEndian.Uint16(byteParser.readBytes(2))
	message.NumAnswers = binary.BigEndian.Uint16(byteParser.readBytes(2))
	message.NumAuthorityRR = binary.BigEndian.Uint16(byteParser.readBytes(2))
	message.NumAdditionalRR = binary.BigEndian.Uint16(byteParser.readBytes(2))
}

func parseQuestions(byteParser *byteParser, message *MDNSMessage) {
	for range message.NumQuestions {
		question := MDNSQuestion{}

		question.Name = parseName(byteParser)
		question.Type = binary.BigEndian.Uint16(byteParser.readBytes(2))
		question.Class = binary.BigEndian.Uint16(byteParser.readBytes(2))

		message.Questions = append(message.Questions, question)
	}
}

func parseAnswers(byteParser *byteParser, message *MDNSMessage) {
	for range message.NumAnswers {
		answer := MDNSAnswer{}

		answer.Name = parseName(byteParser)
		answer.Type = binary.BigEndian.Uint16(byteParser.readBytes(2))
		answer.Class = binary.BigEndian.Uint16(byteParser.readBytes(2))
		answer.TTL = binary.BigEndian.Uint32(byteParser.readBytes(4))
		answer.RDLength = binary.BigEndian.Uint16(byteParser.readBytes(2))

		length := int(answer.RDLength)
		answer.RData = string(byteParser.readBytes(length))

		message.Answers = append(message.Answers, answer)
	}
}

func parseName(byteParser *byteParser) string {
	name := ""

	// bitmask := byte(0b11 << 6)
	// if packet[pos] & bitmask == bitmask {
	// 	// if the first two bits of the NAME word is `11`, then NAME is compressed and we
	// 	// need to parse the remaining 14 bits to get the offset to the first reference of
	// 	// it in another section of the packet
	//
	// 	// TODO: implement
	// }

	for byteParser.peekBytes(1)[0] != '\x00' {
		length := int(byteParser.readBytes(1)[0])
		label := byteParser.readBytes(length)
		name += fmt.Sprintf(".%s", label)
	}

	return name[1:]
}
