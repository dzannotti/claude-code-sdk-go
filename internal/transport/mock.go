package transport

import (
	"context"
	"sync"

	"claudecode/message"
)

type MockTransport struct {
	Connected    bool
	Messages     []message.Message
	ConnectErr   error
	SendErr      error
	InterruptErr error
	CloseErr     error

	SentMessages []StreamMessage
	MsgChan      chan message.Message
	ErrChan      chan error

	mu sync.Mutex
}

func NewMockTransport() *MockTransport {
	return &MockTransport{
		MsgChan: make(chan message.Message, 10),
		ErrChan: make(chan error, 10),
	}
}

func (m *MockTransport) Connect(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ConnectErr != nil {
		return m.ConnectErr
	}

	m.Connected = true

	go func() {
		for _, msg := range m.Messages {
			m.MsgChan <- msg
		}
		close(m.MsgChan)
		close(m.ErrChan)
	}()

	return nil
}

func (m *MockTransport) SendMessage(ctx context.Context, msg StreamMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.SendErr != nil {
		return m.SendErr
	}

	m.SentMessages = append(m.SentMessages, msg)
	return nil
}

func (m *MockTransport) ReceiveMessages(ctx context.Context) (<-chan message.Message, <-chan error) {
	return m.MsgChan, m.ErrChan
}

func (m *MockTransport) Interrupt(ctx context.Context) error {
	return m.InterruptErr
}

func (m *MockTransport) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Connected = false
	return m.CloseErr
}

func (m *MockTransport) IsConnected() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.Connected
}
