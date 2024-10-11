// Package byteparser is a helper module for parsing through a byte slice without
// the need of keeping track of the current byte offset
//
// Useful for parsing packets.
package byteparser

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyInput = errors.New("empty input byte slice")
)

type ByteParser struct {
	input []byte
	offset int
}

func New(input []byte) (*ByteParser, error) {
	if len(input) == 0 {
		return nil, fmt.Errorf("%w: trying to initialize ByteParser with empty input", ErrEmptyInput)
	}

	byteParser := &ByteParser{input: input}

	return byteParser, nil
}

func (bp *ByteParser) ReadBytes(n int) []byte {
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
