package events

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"sync"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
)

func TestMonitorError(t *testing.T) {
	cli := &NopClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := Monitor(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
		// Do nothing
	})

	if err := <-errChan; err == nil {
		t.Fatal("expected an error, got nothing")
	}

}
func TestMonitorErrorDecoding(t *testing.T) {
	cli := &errorEventClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := Monitor(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
		// Do nothing
	})

	if err := <-errChan; err == nil {
		t.Fatal("expected an error, got nothing")
	}

}

type errorEventClient struct {
	NopClient
}

func (c *errorEventClient) Events(ctx context.Context, options types.EventsOptions) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		enc := json.NewEncoder(pw)

		enc.Encode("")

		pw.Close()
	}()

	return ioutil.NopCloser(pr), nil
}

func TestMonitor(t *testing.T) {
	cases := []struct {
		expected []string
		events   []eventtypes.Message
	}{
		{
			expected: []string{},
			events:   []eventtypes.Message{},
		},
		{
			expected: []string{
				"container-create",
			},
			events: []eventtypes.Message{
				{
					Type:   "container",
					Action: "create",
				},
			},
		},
		{
			expected: []string{
				"container-create",
				"network-create",
				"volume-create",
				"container-destroy",
			},
			events: []eventtypes.Message{
				{
					Type:   "container",
					Action: "create",
				},
				{
					Type:   "network",
					Action: "create",
				},
				{
					Type:   "volume",
					Action: "create",
				},
				{
					Type:   "container",
					Action: "destroy",
				},
			},
		},
	}

	for _, c := range cases {
		safeActual := &safeSlice{
			data: []string{},
		}
		cli := &fakeEventClient{
			events: c.events,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		errChan := Monitor(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
			safeActual.Add(m.Type + "-" + m.Action)
		})

		if err := <-errChan; err != nil {
			t.Fatal(err)
		}

		actual := safeActual.Read()
		if !reflect.DeepEqual(c.expected, actual) {
			t.Fatalf("expected %v, got %v", c.expected, actual)
		}
	}
}

func ExampleMonitor() {
	cli, err := client.NewEnvClient()
	if err != nil {
		// Do something..
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errChan := Monitor(ctx, cli, types.EventsOptions{}, func(event eventtypes.Message) {
		fmt.Printf("%v\n", event)
	})

	if err := <-errChan; err != nil {
		// Do something
	}
}

type fakeEventClient struct {
	NopClient
	events []eventtypes.Message
}

func (c *fakeEventClient) Events(ctx context.Context, options types.EventsOptions) (io.ReadCloser, error) {
	pr, pw := io.Pipe()

	go func() {
		enc := json.NewEncoder(pw)

		for _, event := range c.events {
			enc.Encode(event)
			time.Sleep(1 * time.Millisecond)
		}

		pw.Close()
	}()

	return ioutil.NopCloser(pr), nil
}

type safeSlice struct {
	mu   sync.RWMutex
	data []string
}

func (s *safeSlice) Add(element string) {
	s.mu.Lock()
	s.data = append(s.data, element)
	s.mu.Unlock()
}

func (s *safeSlice) Read() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data
}
