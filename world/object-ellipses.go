package world

import (
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

// EllipseObject is a rectangular object
type EllipseObject struct {
	BaseObject

	// radius of EllipseObject
	// b is x radius, a is y radius
	b, a float64
}

// NewEllipseObject return a new rectangular object
func NewEllipseObject(name string, color color.Color, speed, mass, a, b float64, atlas *text.Atlas) *EllipseObject {

	o := &EllipseObject{
		NewBaseObject(name, color, speed, mass, atlas),
		a,
		b,
	}

	return o
}

// BoundingBox returns a Rect, rooted at the center, that covers the object
func (o *EllipseObject) BoundingBox(c pixel.Vec) pixel.Rect {
	min := pixel.V(c.X-o.b, c.Y-o.a)
	max := pixel.V(c.X+o.b, c.Y+o.a)

	return pixel.R(min.X, min.Y, max.X, max.Y)
}

// Draw an ellipse of size width, height inside bounding box set in Phys()
func (o *EllipseObject) Draw(win *pixelgl.Window) {
	if !o.IsSpawned() {
		return
	}

	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	center := o.Phys().Location().Center()

	// mat := pixel.IM

	// o.imd.SetMatrix(mat)
	o.imd.Push(center)
	o.imd.Ellipse(pixel.V(o.b, o.a), 0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// NewEllipseObjectPhys return a new physics object
func NewEllipseObjectPhys(rect pixel.Rect) ObjectPhys {
	return NewBaseObjectPhys(rect)
}