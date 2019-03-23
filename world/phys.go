package world

import (
	"log"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// ObjectPhys is the physics of an object, these values change as the object moves
type ObjectPhys interface {
	Copy() ObjectPhys

	ChangeVerticalDirection(*World)
	CurrentMass() float64
	HandleCollisions(*World) bool
	IsAboveGround(w *World) bool
	Location() pixel.Rect
	OnGround(*World) bool
	Move(*World, pixel.Vec)
	MovingUp() bool
	MovingDown() bool
	MovingLeft() bool
	MovingRight() bool
	PreviousVel() pixel.Vec
	Stopped() bool
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

	parentObject Object
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
func NewBaseObjectPhys(rect pixel.Rect, po Object) ObjectPhys {

	return &BaseObjectPhys{
		rect:         rect,
		parentObject: po, // the object this ObjectPhys belongs to
	}
}

// Copy return a new rectangle phys object based on an existing one
func (o *BaseObjectPhys) Copy() ObjectPhys {

	op := NewBaseObjectPhys(o.Location(), o.parentObject)
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

// OnGround returns true if object is on the ground
func (o *BaseObjectPhys) OnGround(w *World) bool {
	return o.Location().Min.Y == w.Ground.Phys().Location().Max.Y
}

// Stopped returns true if object is stopped
func (o *BaseObjectPhys) Stopped() bool {
	return o.Vel().X == 0
}

// IsAboveGround checks if object is above ground
func (o *BaseObjectPhys) IsAboveGround(w *World) bool {
	return o.Location().Min.Y > w.Ground.Phys().Location().Max.Y
}

// isZeroMass checks if object has no mass
func (o *BaseObjectPhys) isZeroMass() bool {
	return o.CurrentMass() == 0
}

// ChangeVerticalDirection updates the vertical direction if needed
func (o *BaseObjectPhys) ChangeVerticalDirection(w *World) {
	if o.IsAboveGround(w) {
		// fall speed based on mass and gravity
		new := o.Vel()
		new.Y = w.gravity * o.CurrentMass()
		o.SetVel(new)

		if o.Vel().X != 0 {
			v := o.PreviousVel()
			v.X = o.Vel().X
			o.SetPreviousVel(v)

			v = o.Vel()
			v.X = 0
			o.SetVel(v)
		}
	}

	if o.isZeroMass() {
		// rise speed based on mass and gravity
		v := o.Vel()
		v.Y = -1 * w.gravity * o.parentObject.Mass()
		o.SetVel(v)

		if o.Vel().X != 0 {
			v = o.PreviousVel()
			v.X = o.Vel().X
			o.SetPreviousVel(v)

			v = o.Vel()
			v.X = 0
			o.SetVel(v)
		}
	}
}

// MovingUp returns true if object is moving up
func (o *BaseObjectPhys) MovingUp() bool {
	return o.Vel().Y > 0
}

// MovingDown returns true if object is moving down
func (o *BaseObjectPhys) MovingDown() bool {
	return o.Vel().Y < 0
}

// MovingLeft returns true if object is moving left
func (o *BaseObjectPhys) MovingLeft() bool {
	return o.Vel().X < 0
}

// MovingRight returns true if object is moving right
func (o *BaseObjectPhys) MovingRight() bool {
	return o.Vel().X > 0
}

// HandleCollisions returns true if o has any collisions
// it adjusts the physical properties of o to avoid the collision
func (o *BaseObjectPhys) HandleCollisions(w *World) bool {
	switch {
	case o.MovingDown():
		if o.avoidCollisionBelow(w) {
			return true
		}
	case o.MovingUp():
		if o.avoidCollisionAbove(w) {
			return true
		}
	case o.MovingRight():
		if o.avoidCollisionRight(w) {
			return true
		}
	case o.MovingLeft():
		if o.avoidCollisionLeft(w) {
			return true
		}
	}
	return false
}

// shouldCheckVerticalCollision returns true if we need to do a more thourough check of the collision
func (o *BaseObjectPhys) shouldCheckVerticalCollision(other Object) bool {

	if o.parentObject.ID() == other.ID() {
		return false // skip yourself
	}

	if o.Location().Max.X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
		o.Location().Max.X < other.Phys().Location().Min.X {
		return false // no intersection in X axis
	}

	if o.Location().Min.X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
		o.Location().Min.X > other.Phys().Location().Max.X {
		return false // no intersection in X axis
	}
	return true
}

// avoidCollisionBelow changes o to avoid collision with an object below while movign down
func (o *BaseObjectPhys) avoidCollisionBelow(w *World) bool {
	// if about to fall on another, rise back up
	for _, other := range w.CollisionObjects() {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := o.Location().Min.Y - other.Phys().Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Location().Min.Y+o.Vel().Y > other.Phys().Location().Max.Y+other.Phys().Vel().Y &&
			o.Location().Min.Y+o.Vel().Y > other.Phys().Location().Max.Y {
			// too far apart
			continue
		}

		// avoid collision by stopping the fall and rising again
		o.SetCurrentMass(0)
		v := o.Vel()
		v.Y = 0
		o.SetVel(v)
		return true
	}
	return false
}

// avoidCollisionAbove changes o to avoid collision with an object above while moving up
func (o *BaseObjectPhys) avoidCollisionAbove(w *World) bool {
	for _, other := range w.CollisionObjects() {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := other.Phys().Location().Min.Y - o.Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if other.Phys().Location().Min.Y+other.Phys().Vel().Y > o.Location().Max.Y+o.Vel().Y &&
			other.Phys().Location().Min.Y > o.Location().Max.Y+o.Vel().Y {
			// too far apart
			continue
		}

		o.SetCurrentMass(o.parentObject.Mass())
		v := o.Vel()
		v.Y = 0
		o.SetVel(v)
		return true
	}
	return false
}

// shouldCheckHorizontalCollision returns true if we need to do a more thourough check of the collision
func (o *BaseObjectPhys) shouldCheckHorizontalCollision(other Object) bool {

	if o.parentObject.ID() == other.ID() {
		return false // skip yourself
	}

	if other.Phys().Location().Min.Y > o.Location().Max.Y {
		return false // ignore falling BaseObjects higher than you
	}

	return true
}

// ChangeHorizontalDirection changes the horizontal direction of the object to the opposite of current
func (o *BaseObjectPhys) ChangeHorizontalDirection() {
	v := o.Vel()
	v.X = -1 * v.X
	o.SetVel(v)
}

// avoidHorizontalCollision changes the object to avoid a horizontal collision
func (o *BaseObjectPhys) avoidHorizontalCollision() {

	// Going to bump, 50/50 chance of rising up or changing direction
	if utils.RandomInt(0, 100) > 50 {
		o.SetCurrentMass(0)
	} else {
		o.ChangeHorizontalDirection()
	}
}

// avoidCollisionLeft changes o to avoid a collision on the left
func (o *BaseObjectPhys) avoidCollisionLeft(w *World) bool {
	for _, other := range w.CollisionObjects() {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Location().Max.X < other.Phys().Location().Min.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Location().Min.X+o.Vel().X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
			o.Location().Min.X+o.Vel().X > other.Phys().Location().Max.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// avoidCollisionRight changes o to avoid a collision on the right
func (o *BaseObjectPhys) avoidCollisionRight(w *World) bool {
	for _, other := range w.CollisionObjects() {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Location().Min.X > other.Phys().Location().Max.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Location().Max.X+o.Vel().X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
			o.Location().Max.X+o.Vel().X < other.Phys().Location().Min.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// Move moves the object by Vector, accounting for world boundaries
func (o *BaseObjectPhys) Move(w *World, v pixel.Vec) {
	if o.Vel().X != 0 && o.Vel().Y != 0 {
		// cannot currently move in both X and Y direction
		log.Fatalf("o:%+#v\nx: %v; y: %v\n", o, o.Vel().X, o.Vel().Y)
	}

	switch {
	case o.MovingLeft() && o.Location().Min.X+o.Vel().X <= 0:
		// left border
		o.ChangeHorizontalDirection()

	case o.MovingRight() && o.Location().Max.X+o.Vel().X >= w.X:
		// right border
		o.ChangeHorizontalDirection()

	case o.Location().Min.Y+o.Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		o.SetLocation(o.Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-o.Location().Min.Y)))
		v := o.Vel()
		v.Y = 0
		v.X = o.PreviousVel().X
		o.SetVel(v)

	case o.Location().Max.Y+o.Vel().Y >= w.Y && o.Vel().Y > 0:
		// stop at ceiling if going up
		o.SetLocation(o.Location().Moved(pixel.V(0, w.Y-o.Location().Max.Y)))
		v := o.Vel()
		v.Y = 0
		o.SetVel(v)
		o.SetCurrentMass(o.parentObject.Mass())

	default:
		o.SetLocation(o.Location().Moved(pixel.V(v.X, v.Y)))
	}
}
