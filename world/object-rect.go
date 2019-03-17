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

// RectObject is a rectangular object
type RectObject struct {
	BaseObject

	// size of RectObject
	W, H float64
}

// NewRectObject return a new rectangular object
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
	o.phys = phys
	o.Atlas = atlas

	o.phys.SetVel(pixel.V(speed, 0))
	o.phys.SetCurrentMass(mass)

	o.nextPhys = o.phys.Copy()

	return o
}

// Update the RectObject every frame
// o.NextPhys, coming in, is the same as o.Phys.
// Make changes and reads to/from o.NextPhys() only
// When reading properties of other objects, only use other.Phys()
func (o *RectObject) Update(w *World) {
	defer o.CheckIntersect(w)

	// save a copy of the current Phys().object to restore later
	// oldPhys := NewRectObjectPhysCopy(o.phys.(*RectObjectPhys))

	// defer func(o *RectObject) {
	// 	o.nextPhys = o.phys
	// 	o.phys = oldPhys
	// }(o)

	// if on the ground and X velocity is 0, reset it - this seems to be a bug
	if o.NextPhys().Location().Min.Y == w.Ground.Phys().Location().Max.Y && o.NextPhys().Vel().X == 0 {
		v := o.NextPhys().Vel()
		v.X = o.NextPhys().PreviousVel().X
		v.Y = 0
		o.NextPhys().SetVel(v)
	}

	// check if object should rise or fall, these checks not based on collisions
	o.changeVerticalDirection(w)

	// check collisions and adjust movement parameters
	// if a collision is detected, no movement happens this round
	if o.handleCollisions(w) {
		return
	}

	// no collisions detected, move
	o.move(w, pixel.V(o.NextPhys().Vel().X, o.NextPhys().Vel().Y))
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

// RectObjectPhys defines a rectangular object
type RectObjectPhys struct {
	BaseObjectPhys

	// this is the rectangle in the world
	rect pixel.Rect
}

// NewRectObjectPhys return a new physic object
func NewRectObjectPhys(rect pixel.Rect) ObjectPhys {

	return &RectObjectPhys{
		rect: rect,
	}
}

// Copy return a new rectangle phys object based on an existing one
func (o *RectObjectPhys) Copy() ObjectPhys {

	op := NewRectObjectPhys(o.Location())
	op.SetVel(pixel.V(o.Vel().X, o.Vel().Y))
	op.SetPreviousVel(pixel.V(o.PreviousVel().X, o.PreviousVel().Y))
	op.SetCurrentMass(o.CurrentMass())
	return op
}

// Location returns the current location
func (o *RectObjectPhys) Location() pixel.Rect {
	return o.rect
}

// SetLocation sets the current location
func (o *RectObjectPhys) SetLocation(r pixel.Rect) {
	o.rect = r
}