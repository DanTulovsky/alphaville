package world

import (
	"fmt"
	"image/color"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"golang.org/x/image/colornames"
)

// RectObject is an RectObject in the world
type RectObject struct {
	BaseObject

	// size of RectObject
	W, H float64
}

// NewRectObject return a new object in the world
func NewRectObject(name string, color color.Color, speed, mass, W, H float64, phys ObjectPhys, atlas *text.Atlas) *RectObject {
	o := &RectObject{
		W: W,
		H: H,
	}
	o.name = name
	o.id = uuid.New()
	o.color = color
	o.Speed = speed
	o.Mass = mass
	o.imd = imdraw.New(nil)
	o.phys = NewRectObjectPhys()
	o.Atlas = atlas

	return o
}

// ChangeHorizontalDirection changes the horizontal direction of the object to the opposite of current
func (o *RectObject) ChangeHorizontalDirection() {
	v := o.Phys().Vel()
	v.X = -1 * v.X
	o.Phys().SetVel(v)
}

// changeVerticalDirection updates the vertical direction if needed
func (o *RectObject) changeVerticalDirection(w *World) {
	if o.isAboveGround(w) {
		// fall speed based on mass and gravity
		new := o.Phys().Vel()
		new.Y = w.gravity * o.Phys().CurrentMass()
		o.Phys().SetVel(new)

		if o.Phys().Vel().X != 0 {
			v := o.Phys().PreviousVel()
			v.X = o.Phys().Vel().X
			o.Phys().SetPreviousVel(v)

			v = o.Phys().Vel()
			v.X = 0
			o.Phys().SetVel(v)
		}
	}

	if o.isZeroMass() {
		// rise speed based on mass and gravity
		v := o.Phys().Vel()
		v.Y = -1 * w.gravity * o.Mass
		o.Phys().SetVel(v)

		if o.Phys().Vel().X != 0 {
			v = o.Phys().PreviousVel()
			v.X = o.Phys().Vel().X
			o.Phys().SetPreviousVel(v)

			v = o.Phys().Vel()
			v.X = 0
			o.Phys().SetVel(v)
		}
	}
}

// shouldCheckHorizontalCollision returns true if we need to do a more thourough check of the collision
func (o *RectObject) shouldCheckHorizontalCollision(other Object) bool {

	if o.ID() == other.ID() {
		return false // skip yourself
	}

	if other.Phys().Location().Min.Y > o.Phys().Location().Max.Y {
		return false // ignore falling RectObjects higher than you
	}

	return true
}

// shouldCheckVerticalCollision returns true if we need to do a more thourough check of the collision
func (o *RectObject) shouldCheckVerticalCollision(other Object) bool {

	if o.ID() == other.ID() {
		return false // skip yourself
	}

	if o.Phys().Location().Max.X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
		o.Phys().Location().Max.X < other.Phys().Location().Min.X {
		return false // no intersection in X axis
	}

	if o.Phys().Location().Min.X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
		o.Phys().Location().Min.X > other.Phys().Location().Max.X {
		return false // no intersection in X axis
	}
	return true
}

// avoidCollisionBelow changes o to avoid collision with an object below while movign down
func (o *RectObject) avoidCollisionBelow(w *World) bool {
	// if about to fall on another, rise back up
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := o.Phys().Location().Min.Y - other.Phys().Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys().Location().Min.Y+o.Phys().Vel().Y > other.Phys().Location().Max.Y+other.Phys().Vel().Y &&
			o.Phys().Location().Min.Y+o.Phys().Vel().Y > other.Phys().Location().Max.Y {
			// too far apart
			continue
		}

		// avoid collision by stopping the fall and rising again
		o.Phys().SetCurrentMass(0)
		v := o.Phys().Vel()
		v.Y = 0
		o.Phys().SetVel(v)
		return true
	}
	return false
}

// avoidCollisionAbove changes o to avoid collision with an object above while moving up
func (o *RectObject) avoidCollisionAbove(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := other.Phys().Location().Min.Y - o.Phys().Location().Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if other.Phys().Location().Min.Y+other.Phys().Vel().Y > o.Phys().Location().Max.Y+o.Phys().Vel().Y &&
			other.Phys().Location().Min.Y > o.Phys().Location().Max.Y+o.Phys().Vel().Y {
			// too far apart
			continue
		}

		o.Phys().SetCurrentMass(o.Mass)
		v := o.Phys().Vel()
		v.Y = 0
		o.Phys().SetVel(v)
		return true
	}
	return false
}

// avoidHorizontalCollision changes the object to avoid a horizontal collision
func (o *RectObject) avoidHorizontalCollision() {

	// Going to bump, 50/50 chance of rising up or changing direction
	if utils.RandomInt(0, 100) > 50 {
		o.Phys().SetCurrentMass(0)
	} else {
		o.ChangeHorizontalDirection()
	}
}

// avoidCollisionRight changes o to avoid a collision on the right
func (o *RectObject) avoidCollisionRight(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Phys().Location().Min.X > other.Phys().Location().Max.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys().Location().Max.X+o.Phys().Vel().X < other.Phys().Location().Min.X+other.Phys().Vel().X &&
			o.Phys().Location().Max.X+o.Phys().Vel().X < other.Phys().Location().Min.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// avoidCollisionLeft changes o to avoid a collision on the left
func (o *RectObject) avoidCollisionLeft(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Phys().Location().Max.X < other.Phys().Location().Min.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys().Location().Min.X+o.Phys().Vel().X > other.Phys().Location().Max.X+other.Phys().Vel().X &&
			o.Phys().Location().Min.X+o.Phys().Vel().X > other.Phys().Location().Max.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// handleCollisions returns true if o has any collisions
// it adjusts the physical properties of o to avoid the collision
func (o *RectObject) handleCollisions(w *World) bool {
	switch {
	case o.Phys().Vel().Y < 0: // moving down
		if o.avoidCollisionBelow(w) {
			return true
		}
	case o.Phys().Vel().Y > 0: // moving up
		if o.avoidCollisionAbove(w) {
			return true
		}
	case o.Phys().Vel().X > 0: // moving right
		if o.avoidCollisionRight(w) {
			return true
		}
	case o.Phys().Vel().X < 0: // moving left
		if o.avoidCollisionLeft(w) {
			return true
		}
	}
	return false
}

// move moves the object by Vector, accounting for world boundaries
func (o *RectObject) move(w *World, v pixel.Vec) {
	if o.Phys().Vel().X != 0 && o.Phys().Vel().Y != 0 {
		// cannot currently move in both X and Y direction
		log.Fatalf("o:%+#v\nx: %v; y: %v\n", o, o.Phys().Vel().X, o.Phys().Vel().Y)
	}

	switch {
	case o.Phys().Vel().X < 0 && o.Phys().Location().Min.X+o.Phys().Vel().X <= 0:
		// left border
		o.ChangeHorizontalDirection()

	case o.Phys().Vel().X > 0 && o.Phys().Location().Max.X+o.Phys().Vel().X >= w.X:
		// right border
		o.ChangeHorizontalDirection()

	case o.Phys().Location().Min.Y+o.Phys().Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		o.Phys().SetLocation(o.Phys().Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-o.Phys().Location().Min.Y)))
		v := o.Phys().Vel()
		v.Y = 0
		v.X = o.Phys().PreviousVel().X
		o.Phys().SetVel(v)

	case o.Phys().Location().Max.Y+o.Phys().Vel().Y >= w.Y:
		// stop at ceiling
		o.Phys().SetLocation(o.Phys().Location().Moved(pixel.V(0, w.Y-o.Phys().Location().Max.Y)))
		v := o.Phys().Vel()
		v.Y = 0
		o.Phys().SetVel(v)
		o.Phys().SetCurrentMass(o.Mass)

	default:
		o.Phys().SetLocation(o.Phys().Location().Moved(pixel.V(v.X, v.Y)))
	}
}

// Update the RectObject every frame
func (o *RectObject) Update(w *World) {
	defer o.CheckIntersect(w)

	// save a copy of the current Phys().object to restore later
	oldPhys := NewRectObjectPhysCopy(o.Phys())

	defer func(o *RectObject) {
		o.nextPhys = o.phys
		o.phys = oldPhys
	}(o)

	// if on the ground and X velocity is 0, reset it - this seems to be a bug
	if o.Phys().Location().Min.Y == w.Ground.Phys().Location().Max.Y && o.Phys().Vel().X == 0 {
		v := o.Phys().Vel()
		v.X = o.Phys().PreviousVel().X
		v.Y = 0
		o.Phys().SetVel(v)
	}

	// check if object should rise or fall, these checks not based on collisions
	o.changeVerticalDirection(w)

	// check collisions and adjust movement parameters
	// if a collision is detected, no movement happens this round
	if o.handleCollisions(w) {
		return
	}

	// no collisions detected, move
	o.move(w, pixel.V(o.Phys().Vel().X, o.Phys().Vel().Y))
}

// Draw draws the object.
func (o *RectObject) Draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color
	o.imd.Push(o.Phys().Location().Min)
	o.imd.Push(o.Phys().Location().Max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// CheckIntersect prints out an error if this object intersects with another one
func (o *RectObject) CheckIntersect(w *World) {
	for _, other := range w.Objects {
		if o.ID() == other.ID() {
			continue // skip yourself
		}
		if o.Phys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			log.Printf("%#+v (%v) intersects with %#v (%v)", o.name, o.Phys(), other.Name(), other.Phys())
		}
	}
}
