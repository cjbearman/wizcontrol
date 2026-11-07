package bus

import (
	"sync"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

type MessageType string

const (
	bufferSize = 256
)

type Message struct {
	Topic   string
	Payload []byte
}

type Bus struct {
	ch   chan Message
	lock sync.Mutex
}

func NewBus() *Bus {
	return &Bus{
		ch: make(chan Message, bufferSize),
	}
}

func (b *Bus) Publish(msg Message) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.ch != nil {
		b.ch <- msg
	} else {
		common.Logger.GetLogger("bus").Error("Attempted to publish message to closed channel")
	}
}

// Channel returns the channel, which will be nil after close
func (b *Bus) Channel() <-chan Message {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.ch
}

// Close drains the channel and prevents further publishing of messages to it
func (b *Bus) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.ch != nil {
	loop:
		for {
			select {
			case <-b.ch:
			default:
				break loop
			}
		}
		b.ch = nil
	}
}
