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
		CurrentMass: o.CurrentMass,
		Rect:        o.Rect,
	}
}

// Update the Object every frame
func (o *Object) Update(w *World) {
	defer CheckIntersectObject(w, o)

	oldPhys := NewObjectPhysCopy(o.Phys)

	defer func(o *Object) {
		o.NextPhys = o.Phys
		o.Phys = oldPhys
	}(o)

	// if above Ground, fall based on Mass and gravity
	if o.Phys.Rect.Min.Y > w.Ground.Phys.Rect.Max.Y {
		// more Massive Objects fall faster
		o.Phys.Vel.Y = w.gravity * o.Phys.CurrentMass
		if o.Phys.Vel.X != 0 {
			o.Phys.PreviousVel.X = o.Phys.Vel.X // save previous velocity
			o.Phys.Vel.X = 0
		}
	}

	// if Mass is 0, rise based on gravity
	if o.Phys.CurrentMass == 0 {
		o.Phys.Vel.Y = -1 * w.gravity * o.Mass
		if o.Phys.Vel.X != 0 {
			o.Phys.PreviousVel.X = o.Phys.Vel.X // save previous velocity
			o.Phys.Vel.X = 0
		}
	}

	// if on the ground and X velocity is 0, reset it - this seems to be a bug
	if o.Phys.Rect.Min.X == w.Ground.Phys.Rect.Max.X {
		o.Phys.Vel.X = o.Phys.PreviousVel.X
	}

	// falling
	if o.Phys.Vel.Y < 0 {
		// if about to fall on another, rise back up
		for _, other := range w.Objects {
			if o.id == other.id {
				continue // skip yourself
			}

			if o.Phys.Rect.Max.X < other.Phys.Rect.Min.X+other.Phys.Vel.X {
				continue // no intersection in X axis
			}
			if o.Phys.Rect.Min.X > other.Phys.Rect.Max.X+other.Phys.Vel.X {
				continue // no intersection in X axis
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
			return
		}

		// No collisions with other objects detected, check ground.
		if o.Phys.Rect.Min.Y+o.Phys.Vel.Y <= w.Ground.Phys.Rect.Max.Y {
			// stop at ground level
			o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, w.Ground.Phys.Rect.Max.Y-o.Phys.Rect.Min.Y))
			o.Phys.Vel.Y = 0
			o.Phys.Vel.X = o.Phys.PreviousVel.X
		} else {
			o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, o.Phys.Vel.Y))
		}

		return
	}

	// rising
	if o.Phys.Vel.Y > 0 {
		for _, other := range w.Objects {
			if o.id == other.id {
				continue // skip yourself
			}

			if o.Phys.Rect.Max.X < other.Phys.Rect.Min.X || o.Phys.Rect.Min.X > other.Phys.Rect.Max.X {
				continue // no intersection in X axis
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
			return
		}

		// No collision with other objects detected, check ceiling
		if o.Phys.Rect.Max.Y+o.Phys.Vel.Y >= w.Y {
			// would rise above ceiling
			o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, w.Y-o.Phys.Rect.Max.Y))
			o.Phys.Vel.Y = 0
			o.Phys.CurrentMass = o.Mass
		} else {
			o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(0, o.Phys.Vel.Y))
		}
		return
	}

	// move if on the Ground

	// switch directions of at the end of screen, if moving towards end
	if o.Phys.Vel.X < 0 && o.Phys.Rect.Min.X+o.Phys.Vel.X <= 0 {
		o.ChangeDirection()
	}
	if o.Phys.Vel.X > 0 && o.Phys.Rect.Max.X+o.Phys.Vel.X >= w.X {
		o.ChangeDirection()
	}

	// if about to bump into another Object, rise up or change direction
	switch {
	case o.Phys.Vel.X > 0: // moving right
		for _, other := range w.Objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.Phys.Rect.Min.Y > o.Phys.Rect.Max.Y {
				continue // ignore falling Objects higher than you
			}

			if o.Phys.Rect.Min.X > other.Phys.Rect.Max.X {
				continue // no intersection in X axis
			}

			// Check if other moves as expected, or decides to stay in place (due to a third object)
			if o.Phys.Rect.Max.X+o.Phys.Vel.X < other.Phys.Rect.Min.X+other.Phys.Vel.X &&
				o.Phys.Rect.Max.X+o.Phys.Vel.X < other.Phys.Rect.Min.X {
				continue // will not bump
			}

			// Going to bump, 50/50 chance of rising up or changing direction
			if utils.RandomInt(0, 100) > 50 {
				o.Phys.CurrentMass = 0
			} else {
				o.ChangeDirection()
			}
			return
		}
	case o.Phys.Vel.X < 0: // moving left
		for _, other := range w.Objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.Phys.Rect.Min.Y > o.Phys.Rect.Max.Y {
				continue // ignore falling Objects higher than you
			}

			if o.Phys.Rect.Max.X < other.Phys.Rect.Min.X {
				continue // no intersection in X axis
			}

			// Check if other moves as expected, or decides to stay in place (due to a third object)
			if o.Phys.Rect.Min.X+o.Phys.Vel.X > other.Phys.Rect.Max.X+other.Phys.Vel.X &&
				o.Phys.Rect.Min.X+o.Phys.Vel.X > other.Phys.Rect.Max.X {
				continue // will not bump
			}

			// Going to bump, 50/50 chance of rising up or changing direction
			if utils.RandomInt(0, 100) > 50 {
				o.Phys.CurrentMass = 0
			} else {
				o.ChangeDirection()
			}
			return
		}
	}

	// move if nothing else to do
	o.Phys.Rect = o.Phys.Rect.Moved(pixel.V(o.Phys.Vel.X, 0))
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

// CheckIntersectObject prints out an error if this object intersects with another one
func CheckIntersectObject(w *World, o *Object) {
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
func CheckIntersect(w *World) {
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
