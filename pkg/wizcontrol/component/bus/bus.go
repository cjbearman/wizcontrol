// Package bus provides a message bus for dispatching messages to MQTT
package bus

import (
	"sync"

	"github.com/cjbearman/wizcontrol/pkg/wizcontrol/common"
)

const (
	bufferSize = 256
)

// Message represents a message that can be published to the bus, containing a topic and a payload.
type Message struct {
	Topic   string
	Payload []byte
}

// Bus provides a message bus for dispatching messages to MQTT
type Bus struct {
	ch     chan Message
	lock   sync.Mutex
	nilBus bool
}

// NewBus creates a new bus
func NewBus() *Bus {
	return &Bus{
		ch:     make(chan Message, bufferSize),
		nilBus: false,
	}
}

// NewNilBus creates a new bus with a nil channel, effectively disabling message dispatching.
func NewNilBus() *Bus {
	return &Bus{ch: nil, nilBus: true}
}

// Publish sends a message to the MQTT service (if running) via the bus
func (b *Bus) Publish(msg Message) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if b.ch != nil {
		b.ch <- msg
	} else if !b.nilBus {
		common.Logger.GetLogger("bus").Error("Attempted to publish message to closed channel")
	}
}

// Channel returns the channel used for message dispatching, which will be nil after the bus is closed
// or if MQTT is not running
func (b *Bus) Channel() <-chan Message {
	b.lock.Lock()
	defer b.lock.Unlock()
	return b.ch
}

// Close drains the channel and prevents further publishing of messages to it by setting the channel to nil.
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
