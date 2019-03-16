package world

import (
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
)

// CircleObject is a circular object
type CircleObject struct {
	BaseObject

	Radius float64
}

// NewCircleObject returns a new circular object
func NewCircleObject(name string, color color.Color, speed, mass, radius float64, phys ObjectPhys, atlas *text.Atlas) *CircleObject {
	o := &CircleObject{
		Radius: radius,
	}
	o.name = name
	o.id = uuid.New()
	o.color = color
	o.Speed = speed
	o.Mass = mass
	o.imd = imdraw.New(nil)
	o.phys = phys
	o.Atlas = atlas

	return o
}

// Update the CircleObject every frame
func (o *CircleObject) Update(w *World) {
	defer o.CheckIntersect(w)

	// save a copy of the current Phys().object to restore later
	oldPhys := NewCircleObjectPhysCopy(o.phys.(*CircleObjectPhys))

	defer func(o *CircleObject) {
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
func (o *CircleObject) Draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	// Push center point
	v := o.Phys().Location().Min
	// fmt.Printf("v: %#v\n", o.Phys().Location())
	v = v.Add(pixel.V(o.Radius, o.Radius))
	// fmt.Printf("v after: %#v\n", v)
	// fmt.Println()
	o.imd.Push(v)

	// draw circle
	o.imd.Circle(o.Radius, 0)

	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// CircleObjectPhys defines a rectangular object
type CircleObjectPhys struct {
	BaseObjectPhys

	// this is the location of the Object in the world
	circle pixel.Circle
}

// NewCircleObjectPhys return a new physic object
func NewCircleObjectPhys(rect pixel.Rect, radius float64) *CircleObjectPhys {

	// log.Printf("new circle: %#v\n", rect)

	center := pixel.V(rect.Min.X+radius, rect.Min.Y+radius)
	// log.Printf("center: %#v\n", center)

	return &CircleObjectPhys{
		circle: pixel.C(center, radius),
	}
}

// NewCircleObjectPhysCopy return a new rectangle phys object based on an existing one
func NewCircleObjectPhysCopy(o *CircleObjectPhys) *CircleObjectPhys {

	op := NewCircleObjectPhys(o.Location(), o.circle.Radius)
	op.SetVel(pixel.V(o.Vel().X, o.Vel().Y))
	op.SetPreviousVel(pixel.V(o.PreviousVel().X, o.PreviousVel().Y))
	op.SetCurrentMass(o.CurrentMass())
	return op
}

// Circle returns the underlying circle
func (o *CircleObjectPhys) Circle() pixel.Circle {
	return o.circle
}

// Location returns the current location
func (o *CircleObjectPhys) Location() pixel.Rect {

	// fmt.Printf("radius: %v\n", o.circle.Radius)
	// fmt.Printf("new v: %#v\n", pixel.V(o.circle.Radius*-1.0, o.circle.Radius*-1.0))
	min := o.circle.Center.Add(pixel.V(o.circle.Radius*-1, o.circle.Radius*-1))
	max := o.circle.Center.Add(pixel.V(o.circle.Radius, o.circle.Radius))
	b := pixel.R(min.X, min.Y, max.X, max.Y)

	// fmt.Printf("center: %#v\n", o.circle.Center)
	// fmt.Printf("box: %#v\n", b)
	// fmt.Println()
	return b
}

// SetLocation sets the current location
func (o *CircleObjectPhys) SetLocation(r pixel.Rect) {
	center := pixel.V(r.Min.X+o.circle.Radius, r.Min.Y+o.circle.Radius)
	// fmt.Printf("radius2: %#v\n", o.circle.Radius)
	o.circle = pixel.C(center, o.circle.Radius)
}
