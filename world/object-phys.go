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

// RectObjectPhys defines the physical (dynamic) object properties
type RectObjectPhys struct {

	// current horizontal and vertical Speed of Object
	vel pixel.Vec
	// previous horizontal and vertical Speed of Object
	previousVel pixel.Vec

	// currentMass of the Object
	currentMass float64

	// this is the location of the Object in the world
	rect pixel.Rect
}

// NewRectObjectPhys return a new physic object
func NewRectObjectPhys() *RectObjectPhys {
	return &RectObjectPhys{}
}

// NewRectObjectPhysCopy return a new physic object based on an existing one
func NewRectObjectPhysCopy(o ObjectPhys) *RectObjectPhys {

	op := NewRectObjectPhys()
	op.SetVel(pixel.V(o.Vel().X, o.Vel().Y))
	op.SetPreviousVel(pixel.V(o.PreviousVel().X, o.PreviousVel().Y))
	op.SetCurrentMass(o.CurrentMass())
	op.SetLocation(o.Location())
	return op
	// return &RectObjectPhys{
	// 	vel:         pixel.V(o.vel.X, o.vel.Y),
	// 	previousVel: pixel.V(o.previousVel.X, o.previousVel.Y),
	// 	currentMass: o.currentMass,
	// 	rect:        o.rect,
	// }
}

// CurrentMass returns the current mass
func (o *RectObjectPhys) CurrentMass() float64 {
	return o.currentMass
}

// Location returns the current location
func (o *RectObjectPhys) Location() pixel.Rect {
	return o.rect
}

// PreviousVel returns the previous velocity vector
func (o *RectObjectPhys) PreviousVel() pixel.Vec {
	return o.previousVel
}

// Vel returns the current velocity vecotr
func (o *RectObjectPhys) Vel() pixel.Vec {
	return o.vel
}

// SetCurrentMass sets the current mass
func (o *RectObjectPhys) SetCurrentMass(m float64) {
	o.currentMass = m
}

// SetLocation sets the current location
func (o *RectObjectPhys) SetLocation(r pixel.Rect) {
	o.rect = r
}

// SetPreviousVel sets the previous velocity vector
func (o *RectObjectPhys) SetPreviousVel(v pixel.Vec) {
	o.previousVel = v
}

// SetVel sets the current velocity vector
func (o *RectObjectPhys) SetVel(v pixel.Vec) {
	o.vel = v
}
