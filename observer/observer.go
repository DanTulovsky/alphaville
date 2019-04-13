package observer

import (
	"fmt"
	"time"
)

// Event describes an interface for events
type Event interface {
	Description() string
	String() string
	Time() time.Time // time of event

	SetDescription(string)
	SetTime(time.Time)
}

// EventData is a key/value of event specific data
type EventData struct {
	Key   string
	Value string
}

// BaseEvent is the base for all events
type BaseEvent struct {
	data        []EventData
	description string
	time        time.Time // event time
}

// Data returns the event data
func (e *BaseEvent) Data() []EventData {
	return e.data
}

// SetData sets the event data
func (e *BaseEvent) SetData(d []EventData) {
	e.data = d
}

// Description returns the event description
func (e *BaseEvent) Description() string {
	return e.description
}

// SetDescription sets the event description
func (e *BaseEvent) SetDescription(d string) {
	e.description = d
}

// String returns the event as string
func (e *BaseEvent) String() string {
	return fmt.Sprintf("[%v] %v", e.time, e.description)
}

// Time returns the event time
func (e *BaseEvent) Time() time.Time {
	return e.time
}

// SetTime sets the event time
func (e *BaseEvent) SetTime(t time.Time) {
	e.time = t
}

// EventObserver describes an interface for observing events.
type EventObserver interface {
	OnNotify(e Event)
	Name() string
}

// EventNotifier describes an interface for registering and de-registering observers to
// be notified when an event occurs.
type EventNotifier interface {
	Register(e EventObserver)
	Deregister(EventObserver)
	Notify(Event)
}
