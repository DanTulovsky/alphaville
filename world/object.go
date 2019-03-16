package world

import (
	"image/color"

	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
)

// Object is an object in the world
type Object interface {
	Draw(*pixelgl.Window)
	ID() uuid.UUID
	NextPhys() ObjectPhys // returns the NextPhys object
	Name() string
	Phys() ObjectPhys // returns the Phys object
	SwapNextState()
	Update(*World)

	SetNextPhys(ObjectPhys)
	SetPhys(ObjectPhys)
}

// BaseObject is the base object
type BaseObject struct {
	name  string
	id    uuid.UUID
	color color.Color

	// initial Speed and Mass of RectObject
	Speed float64 // horizontal Speed (negative means move left)
	Mass  float64

	// draws the RectObject
	imd *imdraw.IMDraw

	// initial location of the RectObject (bottom left corner)
	IX, IY float64

	// physics properties of the RectObject
	phys     ObjectPhys
	nextPhys ObjectPhys // State of the object in the next round

	Atlas *text.Atlas
}

// Name returns the object's ID
func (o *BaseObject) Name() string {
	return o.name
}

// ID returns the object's ID
func (o *BaseObject) ID() uuid.UUID {
	return o.id
}

// Phys return the phys object
func (o *BaseObject) Phys() ObjectPhys {
	return o.phys
}

// NextPhys return the nextPhys object
func (o *BaseObject) NextPhys() ObjectPhys {
	return o.nextPhys
}

// SetPhys sets the phys object
func (o *BaseObject) SetPhys(op ObjectPhys) {
	o.phys = op
}

// SetNextPhys sets the nextPhys object
func (o *BaseObject) SetNextPhys(op ObjectPhys) {
	o.nextPhys = op
}

// SwapNextState swaps the current state for next state of the object
func (o *BaseObject) SwapNextState() {
	o.phys = o.nextPhys
}

// isAboveGround checks if object is above ground
func (o *BaseObject) isAboveGround(w *World) bool {
	return o.Phys().Location().Min.Y > w.Ground.Phys().Location().Max.Y
}

// isZeroMass checks if object has no mass
func (o *BaseObject) isZeroMass() bool {
	return o.Phys().CurrentMass() == 0
}
