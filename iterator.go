package claudeagent

import (
	"context"
	"sync"

	"claudeagent/message"
)

type MessageIterator interface {
	Next(ctx context.Context) (message.Message, error)
	Close() error
}

type channelIterator struct {
	msgChan <-chan message.Message
	errChan <-chan error
	closeFn func() error
	lastErr error
	closed  bool
	mu      sync.Mutex
}

func newChannelIterator(msgChan <-chan message.Message, errChan <-chan error, closeFn func() error) *channelIterator {
	return &channelIterator{
		msgChan: msgChan,
		errChan: errChan,
		closeFn: closeFn,
	}
}

func (it *channelIterator) Next(ctx context.Context) (message.Message, error) {
	it.mu.Lock()
	if it.closed {
		it.mu.Unlock()
		return nil, ErrAlreadyClosed
	}
	it.mu.Unlock()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg, ok := <-it.msgChan:
			if !ok {
				it.msgChan = nil
				if it.errChan == nil {
					return nil, ErrDone
				}
				continue
			}
			return msg, nil
		case err, ok := <-it.errChan:
			if !ok {
				it.errChan = nil
				if it.msgChan == nil {
					return nil, ErrDone
				}
				continue
			}
			it.lastErr = err
			return nil, err
		}
	}
}

func (it *channelIterator) Close() error {
	it.mu.Lock()
	defer it.mu.Unlock()

	if it.closed {
		return nil
	}
	it.closed = true

	if it.closeFn != nil {
		return it.closeFn()
	}
	return nil
}
