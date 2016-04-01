package events

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"golang.org/x/net/context"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	eventtypes "github.com/docker/engine-api/types/events"
)

func TestMonitorEventsError(t *testing.T) {
	cli := &NopClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := MonitorEvents(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
		// Do nothing
	})

	if err := <-errChan; err == nil {
		t.Fatal("expected an error, got nothing")
	}

}
func TestMonitorEventsErrorDecoding(t *testing.T) {
	cli := &errorEventClient{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	errChan := MonitorEvents(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
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

func TestMonitorEvents(t *testing.T) {
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
		actual := []string{}
		cli := &fakeEventClient{
			events: c.events,
		}

		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		errChan := MonitorEvents(ctx, cli, types.EventsOptions{}, func(m eventtypes.Message) {
			actual = append(actual, m.Type+"-"+m.Action)
		})

		if err := <-errChan; err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(c.expected, actual) {
			t.Fatalf("expected %v, got %v", c.expected, actual)
		}
	}
}

func ExampleMonitorEvents() {
	cli, err := client.NewEnvClient()
	if err != nil {
		// Do something..
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	errChan := MonitorEvents(ctx, cli, types.EventsOptions{}, func(event eventtypes.Message) {
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
