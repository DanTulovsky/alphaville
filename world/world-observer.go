package world

import (
	"log"
	"time"

	"github.com/DanTulovsky/alphaville/observer"
)

func (w *World) processWorldEvent(e *worldEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "none_yet":
			continue
		}
	}
}

func (w *World) processGateEvent(e *GateEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "none_yet":
			continue
		}
	}
}

func (w *World) processTargetEvent(e *TargetEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "destroyed":
			w.RegisterTargetRemoval(data.Value)
		}
	}
}

// OnNotify runs when a notification is received
func (w *World) OnNotify(e observer.Event) {
	switch event := e.(type) {
	case nil:
		log.Printf("nil notification")
	case *worldEvent:
		w.processWorldEvent(event)
	case *GateEvent:
		w.processGateEvent(event)
	case *TargetEvent:
		w.processTargetEvent(event)
	}
}

// Name returns the name of the world
func (w *World) Name() string {
	return "world"
}

type worldEvent struct {
	observer.BaseEvent
}

// NewWorldEvent create a new world event
func (w *World) NewWorldEvent(d string, t time.Time, data ...observer.EventData) observer.Event {
	e := &worldEvent{}
	e.SetData(data)
	e.SetDescription(d)
	e.SetTime(t)

	return e
}

// Implement the observer.EventNotifier interface

// Register registers a new observer for notifying on.
func (w *World) Register(obs observer.EventObserver) {
	w.observers = append(w.observers, obs)
}

// Deregister de-registers an observer for notifying on.
func (w *World) Deregister(obs observer.EventObserver) {
	for i := 0; i < len(w.observers); i++ {
		if obs == w.observers[i] {
			w.observers = append(w.observers[:i], w.observers[i+1:]...)
		}
	}
}

// Notify notifies all observers on an event.
func (w *World) Notify(event observer.Event) {
	for i := 0; i < len(w.observers); i++ {
		w.observers[i].OnNotify(event)
	}
}
