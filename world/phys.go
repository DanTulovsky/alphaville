package world

import "github.com/faiface/pixel"

// ObjectPhys is the physics of an object, these values change as the object moves
type ObjectPhys interface {
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
