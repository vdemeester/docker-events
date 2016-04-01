package events

import (
	"sync"

	eventtypes "github.com/docker/engine-api/types/events"
)

func NewEventHandler(fun func(eventtypes.Message) string) *EventHandler {
	return &EventHandler{
		keyFunc:  fun,
		handlers: make(map[string]func(eventtypes.Message)),
	}
}

func ByType(e eventtypes.Message) string {
	return e.Type
}

func ByAction(e eventtypes.Message) string {
	return e.Action
}

type EventHandler struct {
	keyFunc  func(eventtypes.Message) string
	handlers map[string]func(eventtypes.Message)
	mu       sync.Mutex
}

func (w *EventHandler) Handle(key string, h func(eventtypes.Message)) {
	w.mu.Lock()
	w.handlers[key] = h
	w.mu.Unlock()
}

// Watch ranges over the passed in event chan and processes the events based on the
// handlers created for a given action.
// To stop watching, close the event chan.
func (w *EventHandler) Watch(c <-chan eventtypes.Message) {
	for e := range c {
		w.mu.Lock()
		h, exists := w.handlers[w.keyFunc(e)]
		w.mu.Unlock()
		if !exists {
			continue
		}
		// logrus.Debugf("event handler: received event: %v", e)
		go h(e)
	}
}
