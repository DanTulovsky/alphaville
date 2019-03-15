package world

import "github.com/faiface/pixel"

// RectObjectPhys defines the physical (dynamic) object properties
type RectObjectPhys struct {

	// current horizontal and vertical Speed of Object
	Vel pixel.Vec
	// previous horizontal and vertical Speed of Object
	PreviousVel pixel.Vec

	// currentMass of the Object
	CurrentMass float64

	// this is the location of the Object in the world
	Rect pixel.Rect
}

// NewRectObjectPhys return a new physic object
func NewRectObjectPhys() *RectObjectPhys {
	return &RectObjectPhys{}
}

// NewRectObjectPhysCopy return a new physic object based on an existing one
func NewRectObjectPhysCopy(o *RectObjectPhys) *RectObjectPhys {
	return &RectObjectPhys{
		Vel:         pixel.V(o.Vel.X, o.Vel.Y),
		PreviousVel: pixel.V(o.PreviousVel.X, o.PreviousVel.Y),
		CurrentMass: o.CurrentMass,
		Rect:        o.Rect,
	}
}
