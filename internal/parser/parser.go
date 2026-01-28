package parser

import (
	"fmt"

	"claudeagent/message"
)

const maxBufferSize = 1 << 20 // 1MB

type Parser struct {
	buffer []byte
}

func New() *Parser {
	return &Parser{}
}

func (p *Parser) ProcessLine(line string) ([]message.Message, error) {
	if len(line) == 0 {
		return nil, nil
	}

	if len(line) > maxBufferSize {
		return nil, fmt.Errorf("line exceeds maximum buffer size of %d bytes", maxBufferSize)
	}

	msg, err := message.ParseMessage([]byte(line))
	if err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return []message.Message{msg}, nil
}

func (p *Parser) Reset() {
	p.buffer = p.buffer[:0]
}
