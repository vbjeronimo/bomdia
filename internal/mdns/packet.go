package mdns

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
)

const (
	WORD_SIZE = 2
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

	message := MDNSMessage{Source: source}

	offset := 0
	offset = parseHeader(packet, offset, &message)

	if message.NumQuestions > 0 {
		for _ = range message.NumQuestions {
			offset = parseQuestion(packet, offset, &message)
		}
	}

	if message.NumAnswers > 0 {
		for _ = range message.NumAnswers {
			offset = parseAnswer(packet, offset, &message)
		}
	}

	return &message, nil
}

func parseHeader(packet []byte, offset int, message *MDNSMessage) int {
	message.ID = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE

	flagsSection := packet[offset : offset+WORD_SIZE]

	message.Flags.Query = flagsSection[0]&0x80 != 0
	message.Flags.OPCode = flagsSection[0] >> 3 & 0x0F
	message.Flags.AuthoritativeAnswer = flagsSection[0]&0x04 != 0
	message.Flags.Truncated = flagsSection[0]&0x02 != 0
	message.Flags.RecursionDesired = flagsSection[0]&0x01 != 0

	message.Flags.RecursionAvailable = flagsSection[1]&0x80 != 0
	message.Flags.zero = flagsSection[1] >> 4 & 0x07
	message.Flags.ResponseCode = flagsSection[1] & 0x0F
	offset += WORD_SIZE

	message.NumQuestions = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	message.NumAnswers = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	message.NumAuthorityRR = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	message.NumAdditionalRR = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE

	return offset
}

func parseQuestion(packet []byte, offset int, message *MDNSMessage) int {
	question := MDNSQuestion{}

	question.Name, offset = parseName(packet, offset, message)
	question.Type = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	question.Class = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE

	message.Questions = append(message.Questions, question)

	return offset
}

func parseAnswer(packet []byte, offset int, message *MDNSMessage) int {
	answer := MDNSAnswer{}

	answer.Name, offset = parseName(packet, offset, message)
	answer.Type = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	answer.Class = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE
	answer.TTL = binary.BigEndian.Uint32(packet[offset : offset+(WORD_SIZE*2)])
	offset += WORD_SIZE * 2
	answer.RDLength = binary.BigEndian.Uint16(packet[offset : offset+WORD_SIZE])
	offset += WORD_SIZE

	length := int(answer.RDLength)
	answer.RData = string(packet[offset : offset+length])

	return offset
}

func parseName(packet []byte, offset int, message *MDNSMessage) (string, int) {
	name := ""
	pos := offset

	for packet[pos] != '\x00' {
		length := int(packet[pos])
		pos++

		label := packet[pos : pos+length]
		name += fmt.Sprintf(".%s", label)
		pos += length
	}

	return name[1:], pos
}
