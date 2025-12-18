package claudecode

import (
	"context"
	"errors"
	"testing"
	"time"

	"claudecode/message"
)

func TestChannelIterator_Next_Success(t *testing.T) {
	msgChan := make(chan message.Message, 1)
	errChan := make(chan error, 1)

	msg := &message.UserMessage{
		Type:      "user",
		UUID:      "test-uuid",
		SessionID: "test-session",
	}
	msgChan <- msg
	close(msgChan)
	close(errChan)

	it := newChannelIterator(msgChan, errChan, nil)
	ctx := context.Background()

	result, err := it.Next(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.GetUUID() != "test-uuid" {
		t.Errorf("expected uuid 'test-uuid', got %q", result.GetUUID())
	}
}

func TestChannelIterator_Next_Done(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)
	close(msgChan)
	close(errChan)

	it := newChannelIterator(msgChan, errChan, nil)
	ctx := context.Background()

	_, err := it.Next(ctx)
	if !errors.Is(err, ErrDone) {
		t.Errorf("expected ErrDone, got %v", err)
	}
}

func TestChannelIterator_Next_Error(t *testing.T) {
	msgChan := make(chan message.Message, 1)
	errChan := make(chan error, 1)

	expectedErr := errors.New("test error")
	errChan <- expectedErr
	close(msgChan)
	close(errChan)

	it := newChannelIterator(msgChan, errChan, nil)
	ctx := context.Background()

	_, err := it.Next(ctx)
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestChannelIterator_Next_ContextCanceled(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)

	it := newChannelIterator(msgChan, errChan, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := it.Next(ctx)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestChannelIterator_Next_ContextTimeout(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)

	it := newChannelIterator(msgChan, errChan, nil)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := it.Next(ctx)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded, got %v", err)
	}
}

func TestChannelIterator_Close(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)

	closeCalled := false
	closeFn := func() error {
		closeCalled = true
		return nil
	}

	it := newChannelIterator(msgChan, errChan, closeFn)

	err := it.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !closeCalled {
		t.Error("expected close function to be called")
	}
}

func TestChannelIterator_Close_Idempotent(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)

	callCount := 0
	closeFn := func() error {
		callCount++
		return nil
	}

	it := newChannelIterator(msgChan, errChan, closeFn)

	it.Close()
	it.Close()
	it.Close()

	if callCount != 1 {
		t.Errorf("expected close to be called once, got %d", callCount)
	}
}

func TestChannelIterator_Next_AfterClose(t *testing.T) {
	msgChan := make(chan message.Message)
	errChan := make(chan error)

	it := newChannelIterator(msgChan, errChan, nil)
	it.Close()

	ctx := context.Background()
	_, err := it.Next(ctx)
	if !errors.Is(err, ErrAlreadyClosed) {
		t.Errorf("expected ErrAlreadyClosed, got %v", err)
	}
}
