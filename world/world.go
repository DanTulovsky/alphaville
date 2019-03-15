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
	"golang.org/x/image/font/basicfont"
)

// World defines the world
type World struct {
	X, Y    float64
	Objects []*Object
	Ground  *Object
	gravity float64
	Atlas   *text.Atlas
}

// NewWorld returns a new world
func NewWorld(x, y float64, ground *Object, gravity float64) *World {
	return &World{
		Objects: []*Object{},
		X:       x,
		Y:       y,
		Ground:  ground,
		gravity: gravity,
		Atlas:   text.NewAtlas(basicfont.Face7x13, text.ASCII),
	}
}

// Object is an Object in the world
type Object struct {
	name  string
	id    uuid.UUID
	color color.Color

	// initial Speed and Mass of Object
	Speed float64 // horizontal Speed (negative means move left)
	Mass  float64

	// size of the Object, assuming it's a rectangle
	W, H float64

	// draws the Object
	imd *imdraw.IMDraw

	// initial location of the Object (bottom left corner)
	IX, IY float64

	// physics properties of the Object
	Phys     *ObjectPhys
	NextPhys *ObjectPhys // State of the object in the next round

	Atlas *text.Atlas
}

// NewObject return a new object in the world
func NewObject(name string, color color.Color, speed, mass, W, H float64, phys *ObjectPhys, atlas *text.Atlas) *Object {
	return &Object{
		name:  name,
		id:    uuid.New(),
		color: color,
		Speed: speed,
		Mass:  mass,
		W:     W,
		H:     H,
		imd:   imdraw.New(nil),
		Phys:  phys,
		Atlas: atlas,
	}
}

// ChangeDirection changes the horizontal direction of the object to the opposite of current
func (o *Object) ChangeDirection() {
	o.Phys.Vel.X *= -1
}

// checkFall adjusts the object to fall if it's above ground
func (o *Object) checkFall(w *World) {
	if o.Phys.Rect.Min.Y > w.Ground.Phys.Rect.Max.Y {
		// fall speed based on mass and gravity
		o.Phys.Vel.Y = w.gravity * o.Phys.CurrentMass

		if o.Phys.Vel.X != 0 {
			o.Phys.PreviousVel.X = o.Phys.Vel.X // save previous velocity
			o.Phys.Vel.X = 0
		}
	}
}

// checkRise adjusts the object to rise if needed
func (o *Object) checkRise(w *World) {
	if o.Phys.CurrentMass == 0 {
		// rise speed based on mass and gravity
		o.Phys.Vel.Y = -1 * w.gravity * o.Mass

		if o.Phys.Vel.X != 0 {
			o.Phys.PreviousVel.X = o.Phys.Vel.X // save previous velocity
			o.Phys.Vel.X = 0
		}
	}
}

// shouldCheckHorizontalCollision returns true if we need to do a more thourough check of the collision
func (o *Object) shouldCheckHorizontalCollision(other *Object) bool {

	if o.id == other.id {
		return false // skip yourself
	}

	if other.Phys.Rect.Min.Y > o.Phys.Rect.Max.Y {
		return false // ignore falling Objects higher than you
	}

	return true
}

// shouldCheckVerticalCollision returns true if we need to do a more thourough check of the collision
func (o *Object) shouldCheckVerticalCollision(other *Object) bool {

	if o.id == other.id {
		return false // skip yourself
	}

	if o.Phys.Rect.Max.X < other.Phys.Rect.Min.X+other.Phys.Vel.X &&
		o.Phys.Rect.Max.X < other.Phys.Rect.Min.X {
		return false // no intersection in X axis
	}

	if o.Phys.Rect.Min.X > other.Phys.Rect.Max.X+other.Phys.Vel.X &&
		o.Phys.Rect.Min.X > other.Phys.Rect.Max.X {
		return false // no intersection in X axis
	}
	return true
}

// avoidCollisionBelow changes o to avoid collision with an object below while movign down
func (o *Object) avoidCollisionBelow(w *World) bool {
	// if about to fall on another, rise back up
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := o.Phys.Rect.Min.Y - other.Phys.Rect.Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys.Rect.Min.Y+o.Phys.Vel.Y > other.Phys.Rect.Max.Y+other.Phys.Vel.Y &&
			o.Phys.Rect.Min.Y+o.Phys.Vel.Y > other.Phys.Rect.Max.Y {
			// too far apart
			continue
		}

		// avoid collision by stopping the fall and rising again
		o.Phys.CurrentMass = 0
		o.Phys.Vel.Y = 0
		return true
	}
	return false
}

// avoidCollisionAbove changes o to avoid collision with an object above while moving up
func (o *Object) avoidCollisionAbove(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckVerticalCollision(other) {
			continue
		}

		gap := other.Phys.Rect.Min.Y - o.Phys.Rect.Max.Y
		if gap < 0 {
			continue
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if other.Phys.Rect.Min.Y+other.Phys.Vel.Y > o.Phys.Rect.Max.Y+o.Phys.Vel.Y &&
			other.Phys.Rect.Min.Y > o.Phys.Rect.Max.Y+o.Phys.Vel.Y {
			// too far apart
			continue
		}

		o.Phys.CurrentMass = o.Mass
		o.Phys.Vel.Y = 0
		return true
	}
	return false
}

// avoidHorizontalCollision changes the object to avoid a horizontal collision
func (o *Object) avoidHorizontalCollision() {

	// Going to bump, 50/50 chance of rising up or changing direction
	if utils.RandomInt(0, 100) > 50 {
		o.Phys.CurrentMass = 0
	} else {
		o.ChangeDirection()
	}
}

// avoidCollisionRight changes o to avoid a collision on the right
func (o *Object) avoidCollisionRight(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Phys.Rect.Min.X > other.Phys.Rect.Max.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys.Rect.Max.X+o.Phys.Vel.X < other.Phys.Rect.Min.X+other.Phys.Vel.X &&
			o.Phys.Rect.Max.X+o.Phys.Vel.X < other.Phys.Rect.Min.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// avoidCollisionLeft changes o to avoid a collision on the left
func (o *Object) avoidCollisionLeft(w *World) bool {
	for _, other := range w.Objects {
		if !o.shouldCheckHorizontalCollision(other) {
			continue
		}

		if o.Phys.Rect.Max.X < other.Phys.Rect.Min.X {
			continue // no intersection in X axis
		}

		// Check if other moves as expected, or decides to stay in place (due to a third object)
		if o.Phys.Rect.Min.X+o.Phys.Vel.X > other.Phys.Rect.Max.X+other.Phys.Vel.X &&
			o.Phys.Rect.Min.X+o.Phys.Vel.X > other.Phys.Rect.Max.X {
			continue // will not bump
		}

		o.avoidHorizontalCollision()
		return true
	}
	return false
}

// haveCollisions returns true if o has any collisions
// it adjusts the physical properties of o to avoid the collision
func (o *Object) haveCollision(w *World) bool {
	switch {
	case o.Phys.Vel.Y < 0: // moving down
		if o.avoidCollisionBelow(w) {
			return true
		}
	case o.Phys.Vel.Y > 0: // moving up
		if o.avoidCollisionAbove(w) {
			return true
		}
	case o.Phys.Vel.X > 0: // moving right
		if o.avoidCollisionRight(w) {
			return true
		}
	case o.Phys.Vel.X < 0: // moving left
		if o.avoidCollisionLeft(w) {
			return true
		}
	}
	return false
}

// move moves the object by Vector, accounting for world boundaries
func (o *Object) move(w *World, v pixel.Vec) {
	if o.Phys.Vel.X != 0 && o.Phys.Vel.Y != 0 {
		// cannot currently move in both X and Y direction
		log.Fatalf("o:%+#v\nx: %v; y: %v\n", o, o.Phys.Vel.X, o.Phys.Vel.Y)
	}

	switch {
	case o.Phys.Vel.X < 0 && o.Phys.Rect.Min.X+o.Phys.Vel.X <= 0:
		// left border
		o.ChangeDirection()

	case o.Phys.Vel.X > 0 && o.Phys.Rect.Max.X+o.Phys.Vel.X >= w.X:
		// right border
		o.ChangeDirection()

	case o.Phys.Rect.Min.Y+o.Phys.Vel.Y < w.Ground.Phys.Rect.Max.Y:
		// stop at ground level
		o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, w.Ground.Phys.Rect.Max.Y-o.Phys.Rect.Min.Y))
		o.Phys.Vel.Y = 0
		o.Phys.Vel.X = o.Phys.PreviousVel.X

	case o.Phys.Rect.Max.Y+o.Phys.Vel.Y >= w.Y:
		// stop at ceiling
		o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, w.Y-o.Phys.Rect.Max.Y))
		o.Phys.Vel.Y = 0
		o.Phys.CurrentMass = o.Mass

	default:
		o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(v.X, v.Y))
	}
}

// Update the Object every frame
func (o *Object) Update(w *World) {
	defer o.CheckIntersect(w)

	oldPhys := NewObjectPhysCopy(o.Phys)

	defer func(o *Object) {
		o.NextPhys = o.Phys
		o.Phys = oldPhys
	}(o)

	// if on the ground and X velocity is 0, reset it - this seems to be a bug
	if o.Phys.Rect.Min.Y == w.Ground.Phys.Rect.Max.Y && o.Phys.Vel.X == 0 {
		o.Phys.Vel.X = o.Phys.PreviousVel.X
		o.Phys.Vel.Y = 0
	}

	// check if object should rise or fall, these checks not based on collisions
	o.checkFall(w)
	o.checkRise(w)

	// check collisions and adjust movement parameters
	// if a collision is detected, no movement happens this round
	if o.haveCollision(w) {
		return
	}

	// no collisions detected, move
	o.move(w, pixel.V(o.Phys.Vel.X, o.Phys.Vel.Y))
}

// Draw draws the object.
func (o *Object) Draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color
	o.imd.Push(o.Phys.Rect.Min)
	o.imd.Push(o.Phys.Rect.Max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys.Rect.Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// CheckIntersect prints out an error if this object intersects with another one
func (o *Object) CheckIntersect(w *World) {
	for _, other := range w.Objects {
		if o.id == other.id {
			continue // skip yourself
		}
		if o.Phys.Rect.Intersect(other.Phys.Rect) != pixel.R(0, 0, 0, 0) {
			log.Printf("%#+v (%v) intersects with %#v (%v)", o.name, o.Phys, other.name, other.Phys)
		}
	}
}

// CheckIntersect checks if any objects in the world intersect and prints an error.
func (w *World) CheckIntersect() {
	for _, o := range w.Objects {
		for _, other := range w.Objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if o.Phys.Rect.Intersect(other.Phys.Rect) != pixel.R(0, 0, 0, 0) {
				log.Printf("%#v intersects with %#v", o, other)
			}

		}
	}
}

// ObjectPhys defines the physical (dynamic) object properties
type ObjectPhys struct {

	// current horizontal and vertical Speed of Object
	Vel pixel.Vec
	// previous horizontal and vertical Speed of Object
	PreviousVel pixel.Vec

	// currentMass of the Object
	CurrentMass float64

	// this is the location of the Object in the world
	Rect pixel.Rect
}

// NewObjectPhys return a new physic object
func NewObjectPhys() *ObjectPhys {
	return &ObjectPhys{}
}

// NewObjectPhysCopy return a new physic object based on an existing one
func NewObjectPhysCopy(o *ObjectPhys) *ObjectPhys {
	return &ObjectPhys{
		Vel:         pixel.V(o.Vel.X, o.Vel.Y),
		PreviousVel: pixel.V(o.PreviousVel.X, o.PreviousVel.Y),
		CurrentMass: o.CurrentMass,
		Rect:        o.Rect,
	}
}
