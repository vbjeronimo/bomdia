package mdns

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyInput = errors.New("empty input byte slice")
)

type byteParser struct {
	input []byte
	offset int
}

func newByteParser(input []byte) (*byteParser, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("%w: trying to initialize ByteParser with empty input", ErrEmptyInput)
	}

	byteParser := &byteParser{input: input}

	return byteParser, nil
}

func (bp *byteParser) readBytes(n int) []byte {
	if bp.offset >= len(bp.input) {
		panic("TODO")
	}

	start, end := bp.offset, bp.offset+n

	// TODO: what's the desired behaviour when the parser reaches the
	// end of the input slice? Return '0's, '!ok', 'err', ...?
	b := bp.input[start : end]

	bp.offset = end

	return b
}

func (bp *byteParser) peekBytes(n int) []byte {
	if bp.offset >= len(bp.input) {
		panic("TODO")
	}

	start, end := bp.offset, bp.offset+n
	b := bp.input[start : end]

	return b
}
