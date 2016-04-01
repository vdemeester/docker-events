# docker-events [![Build Status](https://travis-ci.org/vdemeester/docker-events.svg?branch=master)](https://travis-ci.org/vdemeester/docker-events)

A really small library with the intent to ease the use of `Events`
method of `engine-api`.

## Usage

It should be pretty straighforward to use :

```go
import "events"

// […]

cli, err := client.NewEnvClient()
if err != nil {
    // Do something..
}

errChan := events.Monitor(context.Background(), cli, types.EventsOptions{}, func(event eventtypes.Message) {
    fmt.Printf("%v\n", event)
})

if err := <-errChan; err != nil {
    // Do something
}
```

It's also possible to do a little more advanced stuff using
`EventHandler` :

```go
import "events"

// […]

cli, err := client.NewEnvClient()
if err != nil {
    // Do something..
}

// Setup the event handler
eventHandler := events.NewHandler(events.ByAction())
eventHandler.Handle("create", func(m eventtypes.Message) {
    // Do something in case of create message
})

stoppedOrDead := func(m eventtypes.Message) {
    // Do something in case of stop or die message as it might be the
    // same way to react.
}

eventHandler.Handle("die", stoppedOrDead)
eventHandler.Handle("stop", stoppedOrDead)

// The other type of message will be discard.

// Filter the events we wams so receive
filters := filters.NewArgs()
filters.Add("type", "container")
options := types.EventsOptions{
    Filters: filters,
}

errChan := events.MonitorWithHandler(context.Background(), cli, options, eventHandler)

if err := <-errChan; err != nil {
    // Do something
}
```
