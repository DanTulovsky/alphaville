package behavior

import "time"

// Event describes an interface for events
type Event interface {
	Description() string
	String() string
	Time() time.Time // time of event
}

// EventObserver describes an interface for observing events.
type EventObserver interface {
	OnNotify(e Event)
}

// EventNotifier describes an interface for registering and de-registering observers to
// be notified when an event occurs.
type EventNotifier interface {
	Register(EventObserver)
	Deregister(EventObserver)
	Notify(Event)
}

// eventNotifer is an implementation of the EventNotifier interface.
type eventNotifer struct {
	observers []EventObserver
}

// NewEventNotifier returns a new instance of an EventNotifier.
func NewEventNotifier() EventNotifier {
	return &eventNotifer{}
}

// Register registers a new observer for notifying on.
func (e *eventNotifer) Register(obs EventObserver) {
	e.observers = append(e.observers, obs)
}

// Deregister de-registers an observer for notifying on.
func (e *eventNotifer) Deregister(obs EventObserver) {
	for i := 0; i < len(e.observers); i++ {
		if obs == e.observers[i] {
			e.observers = append(e.observers[:i], e.observers[i+1:]...)
		}
	}
}

// Notify notifies all observers on an event.
func (e *eventNotifer) Notify(event Event) {
	for i := 0; i < len(e.observers); i++ {
		e.observers[i].OnNotify(event)
	}
}
