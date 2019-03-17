package world

import (
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

// CircleObject is a rectangular object
type CircleObject struct {
	BaseObject

	// radius of CircleObject
	radius float64
}

// NewCircleObject return a new rectangular object
func NewCircleObject(name string, color color.Color, speed, mass, radius float64, location pixel.Rect, atlas *text.Atlas) *CircleObject {

	phys := NewBaseObjectPhys(location)

	o := &CircleObject{
		NewBaseObject(name, color, speed, mass, phys, atlas),
		radius,
	}

	return o
}

// Draw a rectangle of size width, height inside bounding box set in Phys()
func (o *CircleObject) Draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	center := o.Phys().Location().Center()

	// matrix to manipulate shape
	mat := pixel.IM
	mat = mat.Rotated(center, o.Phys().Angle())

	o.imd.Push(center)
	o.imd.Circle(o.radius, 0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// NewCircleObjectPhys return a new physics object
func NewCircleObjectPhys(rect pixel.Rect) ObjectPhys {
	return NewBaseObjectPhys(rect)
}
