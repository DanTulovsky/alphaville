package world

import (
	"log"
	"time"

	"gogs.wetsnow.com/dant/alphaville/observer"
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
			w.RemoveTarget(data.Value)
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
