package world

import (
	"image/color"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"github.com/DanTulovsky/alphaville/observer"
	"golang.org/x/image/colornames"
)

// NullObject implements the Object interface, but doesn't do anything
type NullObject struct {
	id   uuid.UUID
	name string
}

// NewNullObject returns a new null object
func NewNullObject() *NullObject {
	return &NullObject{
		id: uuid.New(),
	}
}

// BoundingBox always returns 0 rect
func (o *NullObject) BoundingBox(c pixel.Vec) pixel.Rect {
	return pixel.R(0, 0, 0, 0)
}

// Size always returns 0 rect
func (o *NullObject) Size() pixel.Rect {
	return pixel.R(0, 0, 0, 0)
}

// Draw does nothing
func (o *NullObject) Draw(*pixelgl.Window) {}

// Behavior always returns nil
func (o *NullObject) Behavior() Behavior {
	return nil
}

// ID returns the id
func (o *NullObject) ID() uuid.UUID {
	return o.id
}

// Color always return black
func (o *NullObject) Color() color.Color {
	return colornames.Black
}

// IsSpawned always return false
func (o *NullObject) IsSpawned() bool {
	return false
}

// Mass always returns -1
func (o *NullObject) Mass() float64 {
	return -1
}

// NextPhys always returns nil
func (o *NullObject) NextPhys() ObjectPhys {
	return nil
}

// Name always returns 'null'
func (o *NullObject) Name() string {
	return "null"
}

// SetName sets the object's name
func (o *NullObject) SetName(n string) {
	o.name = n
}

// Phys always returns nil
func (o *NullObject) Phys() ObjectPhys {
	return nil
}

// Speed always returns 0
func (o *NullObject) Speed() float64 {
	return 0
}

// SwapNextState does nothing
func (o *NullObject) SwapNextState() {}

// Update does nothing
func (o *NullObject) Update(*World) {}

// SetManualVelocity does nothing
func (o *NullObject) SetManualVelocity(v pixel.Vec) {}

// SetNextPhys does nothing
func (o *NullObject) SetNextPhys(ObjectPhys) {}

// SetPhys does nothing
func (o *NullObject) SetPhys(ObjectPhys) {}

// CheckIntersect does nothing
func (o *NullObject) CheckIntersect(*World) {}

// Implement the observer.EventNotifier interface

// Register does nothing
func (o *NullObject) Register(obs observer.EventObserver) {
}

// Deregister does nothing
func (o *NullObject) Deregister(obs observer.EventObserver) {
}

// Notify does nothing
func (o *NullObject) Notify(event observer.Event) {
}
