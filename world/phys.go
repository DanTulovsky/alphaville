package world

import "github.com/faiface/pixel"

// ObjectPhys is the physics of an object, these values change as the object moves
type ObjectPhys interface {
	Copy() ObjectPhys

	CollisionAbove(*World) bool
	CollisionBelow(*World) bool
	CollisionLeft(*World) bool
	CollisionRight(*World) bool
	CurrentMass() float64
	HaveCollision(*World) bool
	IsAboveGround(w *World) bool
	IsZeroMass() bool
	Location() pixel.Rect
	OnGround(*World) bool
	MoveRight()
	MoveLeft()
	MoveUp()
	MoveDown()
	MovingUp() bool
	MovingDown() bool
	MovingLeft() bool
	MovingRight() bool
	ParentObject() Object
	PreviousVel() pixel.Vec
	Stop()
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

// ParentObject returns the parent object
func (o *BaseObjectPhys) ParentObject() Object {
	return o.parentObject
}

// SetParentObject sets the parent object
func (o *BaseObjectPhys) SetParentObject(po Object) {
	o.parentObject = po
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

// IsZeroMass checks if object has no mass
func (o *BaseObjectPhys) IsZeroMass() bool {
	return o.CurrentMass() == 0
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

// HaveCollision returns true if the object has a collision with current trajectory
func (o *BaseObjectPhys) HaveCollision(w *World) bool {
	return o.CollisionAbove(w) || o.CollisionBelow(w) || o.CollisionLeft(w) || o.CollisionRight(w) || o.CollisionBorders(w)
}

// CollisionBorders returns true if there is a collision with a wall, ground or ceiling
func (o *BaseObjectPhys) CollisionBorders(w *World) bool {

	switch {
	case o.MovingLeft() && o.Location().Min.X+o.Vel().X <= 0:
		// left border
		return true
	case o.MovingRight() && o.Location().Max.X+o.Vel().X >= w.X:
		// right border
		return true
	case o.Location().Min.Y+o.Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		return true
	case o.Location().Max.Y+o.Vel().Y >= w.Y && o.Vel().Y > 0:
		// stop at ceiling if going up
		return true

	default:
		return false
	}
}

// CollisionBelow returns true if object will collide with anything while moving down
func (o *BaseObjectPhys) CollisionBelow(w *World) bool {
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
		return true
	}
	return false
}

// CollisionAbove returns true if object will collide with anything while moveing up
func (o *BaseObjectPhys) CollisionAbove(w *World) bool {
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

	if other.Phys().Location().Max.Y < o.Location().Min.Y {
		return false // ignore objects lower than you
	}

	return true
}

// CollisionLeft returns true if object will collide moving left
func (o *BaseObjectPhys) CollisionLeft(w *World) bool {
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
		return true
	}
	return false
}

// CollisionRight returns true if object will collide moving right
func (o *BaseObjectPhys) CollisionRight(w *World) bool {
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
		return true
	}
	return false
}

// Stop stops the object
func (o *BaseObjectPhys) Stop() {
	v := o.Vel()
	v.X = 0
	v.Y = 0
	o.SetVel(v)
}

// MoveLeft moves the object left
func (o *BaseObjectPhys) MoveLeft() {
	v := o.Vel()
	v.X = o.ParentObject().Speed() * -1
	v.Y = 0
	o.SetVel(v)
}

// MoveRight moves the object right
func (o *BaseObjectPhys) MoveRight() {
	v := o.Vel()
	v.X = o.ParentObject().Speed()
	v.Y = 0
	o.SetVel(v)
}

// MoveUp moves the object up
func (o *BaseObjectPhys) MoveUp() {
	v := o.Vel()
	v.Y = o.ParentObject().Speed()
	v.X = 0
	o.SetVel(v)
}

// MoveDown moves the object down
func (o *BaseObjectPhys) MoveDown() {
	v := o.Vel()
	v.Y = o.ParentObject().Speed() * -1
	v.X = 0
	o.SetVel(v)
}
