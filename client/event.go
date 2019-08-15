package client

import (
	"sync"
)

var clientID int32 // store client.GetMe().id once Collector in created (to be used with FilterMe and FilterNotMe)

// Collector provides the instance for collecting events (updates)
type Collector struct {
	client           *Client
	wg               sync.WaitGroup
	registerListener chan *Listener
	listeners        []*Listener
	sync.Mutex
}

//NewEventCollector - creates a new Collector instance with default configuration
func NewEventCollector(c *Client) *Collector {
	me, _ := c.GetMe()
	clientID = me.Id
	return &Collector{client: c, registerListener: make(chan *Listener, 1)}
}

func (collector *Collector) getMessage(update Type) Message {
	switch update.GetType() {
	case TypeUpdateNewMessage:
		return *update.(*UpdateNewMessage).Message
	default:
		return Message{}
	}
}

// OnMessage for incoming and outgoing messages that don't have a request cost
func (collector *Collector) OnMessage(fn func(*Client, Message), filters ...string) {
	listener := collector.client.GetListener()
	collector.registerListener <- listener
	for update := range listener.Updates {
		if update.GetClass() == ClassUpdate {
			message := collector.getMessage(update)
			if (Message{}) != message && applyFilters(&message, filters) {
				fn(collector.client, message)
			}
		}

	}
}

func (collector *Collector) getEditedMessage(update Type) Message {
	if update.GetType() == TypeUpdateMessageEdited {
		info := *update.(*UpdateMessageEdited)
		message, err := collector.client.GetMessage(
			&GetMessageRequest{
				ChatId:    info.ChatId,
				MessageId: info.MessageId,
			})
		if err == nil {
			return *message
		}
	}
	return Message{}
}

// OnEditedMessage there's a request cost in receiving edited messages,
// therefore it's a separate method (requires a get request that may fail for varius reasons)
func (collector *Collector) OnEditedMessage(fn func(*Client, Message), filters ...string) {
	filters = append(filters, "FilterEdited")
	listener := collector.client.GetListener()
	collector.registerListener <- listener
	for update := range listener.Updates {
		if update.GetClass() == ClassUpdate {
			message := collector.getEditedMessage(update)
			if (Message{}) != message && applyFilters(&message, filters) {
				fn(collector.client, message)
			}
		}

	}
}

// OnCommand commands are a specific type of filter on incoming messages that are sent by others
func (collector *Collector) OnCommand(fn func(*Client, []string), cmd, prefix string) {
	filters := []string{"FilterIncoming", "FilterNotMe", "FilterText"}
	listener := collector.client.GetListener()
	collector.registerListener <- listener
	for update := range listener.Updates {
		if update.GetClass() == ClassUpdate {
			message := collector.getMessage(update)
			if (Message{}) != message && applyFilters(&message, filters) {
				if args := isValidCommand(&message, cmd, prefix); args != nil {
					fn(collector.client, args)
				}
			}
		}

	}
}

// Wait - Wait on events
func (collector *Collector) Wait() {
	go func(ch chan *Listener) {
		for value := range ch {
			collector.Lock()
			collector.listeners = append(collector.listeners, value)
			collector.Unlock()
		}
	}(collector.registerListener)

	collector.wg.Add(1)
	collector.wg.Wait()
}

//Close - closes Collector instance
func (collector *Collector) Close() {
	collector.Lock()
	listeners := collector.listeners
	collector.Unlock()
	for _, item := range listeners {
		item.Close()
	}
	collector.wg.Done()
}
