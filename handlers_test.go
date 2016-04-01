package events

import (
	"reflect"
	"testing"
	"time"

	eventtypes "github.com/docker/engine-api/types/events"
)

func TestByType(t *testing.T) {
	cases := []struct {
		message  eventtypes.Message
		expected string
	}{
		{
			message:  eventtypes.Message{},
			expected: "",
		},
		{
			message: eventtypes.Message{
				Type: "container",
			},
			expected: "container",
		},
		{
			message: eventtypes.Message{
				Type: "image",
			},
			expected: "image",
		},
	}
	for _, c := range cases {
		actual := ByType(c.message)
		if actual != c.expected {
			t.Fatalf("expected %s, got %s", c.expected, actual)
		}
	}
}

func TestAction(t *testing.T) {
	cases := []struct {
		message  eventtypes.Message
		expected string
	}{
		{
			message:  eventtypes.Message{},
			expected: "",
		},
		{
			message: eventtypes.Message{
				Action: "start",
			},
			expected: "start",
		},
		{
			message: eventtypes.Message{
				Action: "die",
			},
			expected: "die",
		},
	}
	for _, c := range cases {
		actual := ByAction(c.message)
		if actual != c.expected {
			t.Fatalf("expected %s, got %s", c.expected, actual)
		}
	}
}

func TestWatchNoFiltering(t *testing.T) {
	safeActual := &safeSlice{
		data: []string{},
	}
	expectedEvents := []string{
		"container-create",
		"container-start",
		"network-create",
	}
	eventChan := make(chan eventtypes.Message)

	go func() {
		eventChan <- eventtypes.Message{
			Type:   "container",
			Action: "create",
		}
		time.Sleep(1 * time.Millisecond)
		eventChan <- eventtypes.Message{
			Type:   "container",
			Action: "start",
		}
		time.Sleep(1 * time.Millisecond)
		eventChan <- eventtypes.Message{
			Type:   "network",
			Action: "create",
		}
		time.Sleep(1 * time.Millisecond)
		close(eventChan)
	}()

	h := NewHandler(func(e eventtypes.Message) string { return "" })
	h.Handle("", func(e eventtypes.Message) {
		safeActual.Add(e.Type + "-" + e.Action)
	})
	h.Watch(eventChan)

	actualEvents := safeActual.Read()
	if !reflect.DeepEqual(actualEvents, expectedEvents) {
		t.Fatalf("expected %v, got %v", expectedEvents, actualEvents)
	}
}

func TestWatchFiltering(t *testing.T) {
	safeActual := &safeSlice{
		data: []string{},
	}
	expectedEvents := []string{
		"container-create",
		"container-start",
	}
	eventChan := make(chan eventtypes.Message)

	go func() {
		eventChan <- eventtypes.Message{
			Type:   "container",
			Action: "create",
		}
		time.Sleep(1 * time.Millisecond)
		eventChan <- eventtypes.Message{
			Type:   "container",
			Action: "start",
		}
		time.Sleep(1 * time.Millisecond)
		eventChan <- eventtypes.Message{
			Type:   "network",
			Action: "create",
		}
		time.Sleep(1 * time.Millisecond)
		close(eventChan)
	}()

	h := NewHandler(func(e eventtypes.Message) string { return e.Type })
	h.Handle("container", func(e eventtypes.Message) {
		safeActual.Add(e.Type + "-" + e.Action)
	})
	h.Watch(eventChan)

	actualEvents := safeActual.Read()
	if !reflect.DeepEqual(actualEvents, expectedEvents) {
		t.Fatalf("expected %v, got %v", expectedEvents, actualEvents)
	}
}
