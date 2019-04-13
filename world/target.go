package world

import (
	"bytes"
	"html/template"
	"image/color"
	"log"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"golang.org/x/image/colornames"
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
	// observer.EventNotifier
	Object

	Available() bool
	SetAvailable(bool)
	Bounds() pixel.Rect
	Circle() pixel.Circle
	Destroy()
	Location() pixel.Vec
}

type simpleTarget struct {
	BaseObject
	id        uuid.UUID
	name      string
	color     color.Color
	observers []observer.EventObserver
	bounds    pixel.Rect // bounding box of target
	available bool
}

// NewSimpleTarget returns a new simple target
func NewSimpleTarget(name string, l pixel.Vec, r float64, d string) Target {
	st := &simpleTarget{
		id:        uuid.New(),
		name:      name,
		bounds:    pixel.R(l.X-r, l.Y-r, l.X+r, l.Y+r),
		color:     colornames.Red,
		available: true,
	}
	st.description = d
	return st
}

// String ...
func (t *simpleTarget) String() string {

	buf := bytes.NewBufferString("")
	tmpl, err := template.New("simpleTarget").Parse(
		`
Target
  Name: {{.Name}}	
	Desc: {{.Description}}	
	Available: {{.Available}}
	Location: {{.Circle.Center}} (r={{.Circle.Radius}})
	ID: {{.ID}}
`)

	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}
	err = tmpl.Execute(buf, t)
	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}

	return buf.String()
}

// Available returns availability of the target
func (t *simpleTarget) Available() bool {
	return t.available
}

// SetAvailable sets availability
func (t *simpleTarget) SetAvailable(a bool) {
	t.available = a
}

// Color returns the color of the target
func (t *simpleTarget) Color() color.Color {
	return t.color
}

// Bounds returns the bounds of the target
func (t *simpleTarget) Bounds() pixel.Rect {
	return t.bounds
}

// Circle returns the underlying circle of the target
func (t *simpleTarget) Circle() pixel.Circle {
	return pixel.C(t.bounds.Center(), t.bounds.W()/2)
}

// ID returns the id of the target
func (t *simpleTarget) ID() uuid.UUID {
	return t.id
}

// Implement the observer.EventNotifier interface

// Register registers a new observer for notifying on.
func (t *simpleTarget) Register(obs observer.EventObserver) {
	t.observers = append(t.observers, obs)
}

// Deregister de-registers an observer for notifying on.
// TODO: This causes problems if called because it modifies the t.observers list
// that something is iterating over.
func (t *simpleTarget) Deregister(obs observer.EventObserver) {
	for i := 0; i < len(t.observers); i++ {
		if obs == t.observers[i] {
			t.observers = append(t.observers[:i], t.observers[i+1:]...)
		}
	}
}

// Notify notifies all observers on an event.
func (t *simpleTarget) Notify(event observer.Event) {
	// t.observers gets modified by objects unregistering on destruction
	observers := t.observers
	// log.Printf("observers before: %v", len(observers))
	for i := 0; i < len(observers); i++ {
		// log.Printf("processing observer %v", observers[i].Name())
		t.observers[i].OnNotify(event)
	}
	// log.Printf("observers after: %v", len(observers))
}

// Location returns the target's location
func (t *simpleTarget) Location() pixel.Vec {
	return t.bounds.Center()
}

func (t *simpleTarget) Name() string {
	return t.name
}

// Destroy destroys this target
// A notification is issued and the world is updated via it
func (t *simpleTarget) Destroy() {
	t.Notify(NewTargetEvent(
		"target destroyed", time.Now(), observer.EventData{Key: "destroyed", Value: t.id.String()}))
	// t = nil
}

// Draw draws the target
func (t *simpleTarget) Draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = t.color
	imd.Push(t.Location())
	imd.Circle(t.bounds.W()/2, 0)
	imd.Draw(win)
}
