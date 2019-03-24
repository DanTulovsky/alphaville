package world

import (
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/observer"
)

// TargetEvent implements the observer.Event interface to send events to other components
type TargetEvent struct {
	observer.BaseEvent
}

// NewTargetEvent create a new target event
func NewTargetEvent(d string, t time.Time, data ...observer.EventData) observer.Event {
	e := &TargetEvent{}
	e.SetData(data)
	e.SetDescription(d)
	e.SetTime(t)

	return e
}

// Target is something that target seekers hunt
type Target interface {
	ID() uuid.UUID
	Destroy()
	Draw(*pixelgl.Window)
	EventNotifier() observer.EventNotifier
	Location() pixel.Vec
	Name() string
}

type simpleTarget struct {
	id            uuid.UUID
	name          string
	location      pixel.Vec
	eventNotifier observer.EventNotifier
}

// NewSimpleTarget returns a new simple target
func NewSimpleTarget(name string, l pixel.Vec) Target {
	return &simpleTarget{
		id:            uuid.New(),
		name:          name,
		location:      l,
		eventNotifier: observer.NewEventNotifier(),
	}
}

// ID returns the id of the target
func (t *simpleTarget) ID() uuid.UUID {
	return t.id
}

// EventNotifier returns the event notifier
func (t *simpleTarget) EventNotifier() observer.EventNotifier {
	return t.eventNotifier
}

// Location returns the target's location
func (t *simpleTarget) Location() pixel.Vec {
	return t.location
}

func (t *simpleTarget) Name() string {
	return t.name
}

// Destroy destroys this target
// A notification is issued and the world is updated via it
func (t *simpleTarget) Destroy() {
	t.eventNotifier.Notify(NewTargetEvent(
		"target destroyed", time.Now(), observer.EventData{Key: "destroyed", Value: t.id.String()}))
}

// Draw draws the target
func (t *simpleTarget) Draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.Red
	imd.Push(t.Location())
	imd.Circle(10, 0)
	imd.Draw(win)
}
