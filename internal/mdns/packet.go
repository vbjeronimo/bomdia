package mdns

import "net"

type MDNSPacket struct {
	Source net.UDPAddr

	ID uint16
	Flags MDNSFlags
	NumQuestions uint16
	NumAnswers uint16
	NumAuthorityRR uint16
	NumAdditionalRR uint16

	Questions []MDNSQuestion
	Answers []MDNSAnswer

	// Not sure yet if I'll need the other two slices, so I'll skip them for now
}

type MDNSFlags struct {
	Query bool
	AuthoritativeAnswer bool
	Truncated bool
	RecursionDesired bool
	RecursionAvailable bool

	OPCode uint8
	ResponseCode uint8

	zero uint8 // The "Zero bits" are a reserved section that's not used for anything, so I won't be exporting it
}

type MDNSQuestion struct {
	Name string
	Type uint16
	Class uint16
}

type MDNSAnswer struct {
	Name string
	Type uint16
	Class uint16
	TTL uint32
	RDLength uint16
	RData string
}
