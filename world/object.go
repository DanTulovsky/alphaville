package world

import (
	"image/color"
	"log"

	"gogs.wetsnow.com/dant/alphaville/utils"

	"github.com/faiface/pixel"
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

	// initial Speed and Mass of BaseObject
	Speed float64 // horizontal Speed (negative means move left)
	Mass  float64

	// draws the BaseObject
	imd *imdraw.IMDraw

	// initial location of the BaseObject (bottom left corner)
	IX, IY float64

	// physics properties of the BaseObject
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
	// o.phys = o.nextPhys
	o.phys = o.nextPhys.Copy()
}

// isAboveGround checks if object is above ground
func (o *BaseObject) isAboveGround(w *World) bool {
	return o.NextPhys().Location().Min.Y > w.Ground.Phys().Location().Max.Y
}

// isZeroMass checks if object has no mass
func (o *BaseObject) isZeroMass() bool {
	return o.NextPhys().CurrentMass() == 0
}

// ChangeHorizontalDirection changes the horizontal direction of the object to the opposite of current
func (o *BaseObject) ChangeHorizontalDirection() {
	v := o.NextPhys().Vel()
	v.X = -1 * v.X
	o.NextPhys().SetVel(v)
}

// changeVerticalDirection updates the vertical direction if needed
func (o *BaseObject) changeVerticalDirection(w *World) {
	if o.isAboveGround(w) {
		// fall speed based on mass and gravity
		new := o.NextPhys().Vel()
		new.Y = w.gravity * o.NextPhys().CurrentMass()
		o.NextPhys().SetVel(new)

		if o.NextPhys().Vel().X != 0 {
			v := o.NextPhys().PreviousVel()
			v.X = o.NextPhys().Vel().X
			o.NextPhys().SetPreviousVel(v)

			v = o.NextPhys().Vel()
			v.X = 0
			o.NextPhys().SetVel(v)
		}
	}

	if o.isZeroMass() {
		// rise speed based on mass and gravity
		v := o.NextPhys().Vel()
		v.Y = -1 * w.gravity * o.Mass
		o.NextPhys().SetVel(v)

		if o.NextPhys().Vel().X != 0 {
			v = o.NextPhys().PreviousVel()
			v.X = o.NextPhys().Vel().X
			o.NextPhys().SetPreviousVel(v)

			v = o.NextPhys().Vel()
			v.X = 0
			o.NextPhys().SetVel(v)
		}
	}
}

// shouldCheckHorizontalCollision returns true if we need to do a more thourough check of the collision
func (o *BaseObject) shouldCheckHorizontalCollision(other Object) bool {

	if o.ID() == other.ID() {
		return false // skip yourself
	}

	if other.Phys().Location().Min.Y > o.NextPhys().Location().Max.Y {
		return false // ignore falling BaseObjects higher than you
	}

	return true
}

// shouldCheckVerticalCollision returns true if we need to do a more thourough check of the collision
func (o *BaseObject) shouldCheckVerticalCollision(other Object) bool {

	if o.ID() == other.ID() {
		return false // skip yourself
	}

	if o.NextPhys().Location().Max.X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
		o.NextPhys().Location().Max.X < other.Phys().Location().Min.X {
		return false // no intersection in X axis
	}

	if o.NextPhys().Location().Min.X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
		o.NextPhys().Location().Min.X > other.Phys().Location().Max.X {
		return false // no intersection in X axis
	}
	return true
}

// avoidCollisionBelow changes o to avoid collision with an object below while movign down
func (o *BaseObject) avoidCollisionBelow(w *World) bool {
	// if about to fall on another, rise back up
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := o.NextPhys().Location().Min.Y - other.Phys().Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.NextPhys().Location().Min.Y+o.NextPhys().Vel().Y > other.Phys().Location().Max.Y+other.Phys().Vel().Y &&
			o.NextPhys().Location().Min.Y+o.NextPhys().Vel().Y > other.Phys().Location().Max.Y {
			// too far apart
			continue
		}

		// avoid collision by stopping the fall and rising again
		o.NextPhys().SetCurrentMass(0)
		v := o.NextPhys().Vel()
		v.Y = 0
		o.NextPhys().SetVel(v)
		return true
	}
	return false
}

// avoidCollisionAbove changes o to avoid collision with an object above while moving up
func (o *BaseObject) avoidCollisionAbove(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := other.Phys().Location().Min.Y - o.NextPhys().Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if other.Phys().Location().Min.Y+other.Phys().Vel().Y > o.NextPhys().Location().Max.Y+o.NextPhys().Vel().Y &&
			other.Phys().Location().Min.Y > o.NextPhys().Location().Max.Y+o.NextPhys().Vel().Y {
			// too far apart
			continue
		}

		o.NextPhys().SetCurrentMass(o.Mass)
		v := o.NextPhys().Vel()
		v.Y = 0
		o.NextPhys().SetVel(v)
		return true
	}
	return false
}

// avoidHorizontalCollision changes the object to avoid a horizontal collision
func (o *BaseObject) avoidHorizontalCollision() {

	// Going to bump, 50/50 chance of rising up or changing direction
	if utils.RandomInt(0, 100) > 50 {
		o.NextPhys().SetCurrentMass(0)
	} else {
		o.ChangeHorizontalDirection()
	}
}

// avoidCollisionRight changes o to avoid a collision on the right
func (o *BaseObject) avoidCollisionRight(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.NextPhys().Location().Min.X > other.Phys().Location().Max.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.NextPhys().Location().Max.X+o.NextPhys().Vel().X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
			o.NextPhys().Location().Max.X+o.NextPhys().Vel().X < other.Phys().Location().Min.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// avoidCollisionLeft changes o to avoid a collision on the left
func (o *BaseObject) avoidCollisionLeft(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.NextPhys().Location().Max.X < other.Phys().Location().Min.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.NextPhys().Location().Min.X+o.NextPhys().Vel().X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
			o.NextPhys().Location().Min.X+o.NextPhys().Vel().X > other.Phys().Location().Max.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// handleCollisions returns true if o has any collisions
// it adjusts the physical properties of o to avoid the collision
func (o *BaseObject) handleCollisions(w *World) bool {
	switch {
	case o.NextPhys().Vel().Y < 0: // moving down
		if o.avoidCollisionBelow(w) {
			return true
		}
	case o.NextPhys().Vel().Y > 0: // moving up
		if o.avoidCollisionAbove(w) {
			return true
		}
	case o.NextPhys().Vel().X > 0: // moving right
		if o.avoidCollisionRight(w) {
			return true
		}
	case o.NextPhys().Vel().X < 0: // moving left
		if o.avoidCollisionLeft(w) {
			return true
		}
	}
	return false
}

// move moves the object by Vector, accounting for world boundaries
func (o *BaseObject) move(w *World, v pixel.Vec) {
	if o.NextPhys().Vel().X != 0 && o.NextPhys().Vel().Y != 0 {
		// cannot currently move in both X and Y direction
		log.Fatalf("o:%+#v\nx: %v; y: %v\n", o, o.NextPhys().Vel().X, o.NextPhys().Vel().Y)
	}

	switch {
	case o.NextPhys().Vel().X < 0 && o.NextPhys().Location().Min.X+o.NextPhys().Vel().X <= 0:
		// left border
		o.ChangeHorizontalDirection()

	case o.NextPhys().Vel().X > 0 && o.NextPhys().Location().Max.X+o.NextPhys().Vel().X >= w.X:
		// right border
		o.ChangeHorizontalDirection()

	case o.NextPhys().Location().Min.Y+o.NextPhys().Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		o.NextPhys().SetLocation(o.NextPhys().Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-o.NextPhys().Location().Min.Y)))
		v := o.NextPhys().Vel()
		v.Y = 0
		v.X = o.NextPhys().PreviousVel().X
		o.NextPhys().SetVel(v)

	case o.NextPhys().Location().Max.Y+o.NextPhys().Vel().Y >= w.Y && o.NextPhys().Vel().Y > 0:
		// stop at ceiling if going up
		o.NextPhys().SetLocation(o.NextPhys().Location().Moved(pixel.V(0, w.Y-o.NextPhys().Location().Max.Y)))
		v := o.NextPhys().Vel()
		v.Y = 0
		o.NextPhys().SetVel(v)
		o.NextPhys().SetCurrentMass(o.Mass)

	default:
		o.NextPhys().SetLocation(o.NextPhys().Location().Moved(pixel.V(v.X, v.Y)))
	}
}

// CheckIntersect prints out an error if this object intersects with another one
func (o *BaseObject) CheckIntersect(w *World) {
	for _, other := range w.Objects {
		if o.ID() == other.ID() {
			continue // skip yourself
		}
		if o.NextPhys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			log.Printf("%#+v (%v) intersects with %#v (%v)", o.name, o.NextPhys(), other.Name(), other.Phys())
		}
	}
}
