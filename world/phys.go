package world

import "github.com/faiface/pixel"

// ObjectPhys is the physics of an object, these values change as the object moves
type ObjectPhys interface {
	Copy() ObjectPhys

	CurrentMass() float64
	Location() pixel.Rect
	PreviousVel() pixel.Vec
	Vel() pixel.Vec

	SetCurrentMass(float64)
	SetLocation(pixel.Rect)
	SetPreviousVel(pixel.Vec)
	SetVel(pixel.Vec)
}

// BaseObjectPhys defines the physical (dynamic) object properties
type BaseObjectPhys struct {

	// current horizontal and vertical Speed of Object
	vel pixel.Vec

	// previous horizontal and vertical Speed of Object
	previousVel pixel.Vec

	// currentMass of the Object
	currentMass float64

	// this is the bounding rectangle in the world
	rect pixel.Rect
}

// CurrentMass returns the current mass
func (o *BaseObjectPhys) CurrentMass() float64 {
	return o.currentMass
}

// PreviousVel returns the previous velocity vector
func (o *BaseObjectPhys) PreviousVel() pixel.Vec {
	return o.previousVel
}

// Vel returns the current velocity vecotr
func (o *BaseObjectPhys) Vel() pixel.Vec {
	return o.vel
}

// SetCurrentMass sets the current mass
func (o *BaseObjectPhys) SetCurrentMass(m float64) {
	o.currentMass = m
}

// SetPreviousVel sets the previous velocity vector
func (o *BaseObjectPhys) SetPreviousVel(v pixel.Vec) {
	o.previousVel = v
}

// SetVel sets the current velocity vector
func (o *BaseObjectPhys) SetVel(v pixel.Vec) {
	o.vel = v
}

// NewBaseObjectPhys return a new physic object
func NewBaseObjectPhys(rect pixel.Rect) ObjectPhys {

	return &BaseObjectPhys{
		rect: rect,
	}
}

// Copy return a new rectangle phys object based on an existing one
func (o *BaseObjectPhys) Copy() ObjectPhys {

	op := NewBaseObjectPhys(o.Location())
	op.SetVel(pixel.V(o.Vel().X, o.Vel().Y))
	op.SetPreviousVel(pixel.V(o.PreviousVel().X, o.PreviousVel().Y))
	op.SetCurrentMass(o.CurrentMass())
	return op
}

// Location returns the current location
func (o *BaseObjectPhys) Location() pixel.Rect {
	return o.rect
}

// SetLocation sets the current location
func (o *BaseObjectPhys) SetLocation(r pixel.Rect) {
	o.rect = r
}
