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
	// b is y radius, a is x radius
	b, a float64
}

// NewEllipseObject return a new rectangular object
func NewEllipseObject(name string, color color.Color, speed, mass, a, b float64, location pixel.Rect, atlas *text.Atlas) *EllipseObject {

	phys := NewBaseObjectPhys(location)

	o := &EllipseObject{
		NewBaseObject(name, color, speed, mass, phys, atlas),
		a,
		b,
	}

	return o
}

// Draw an ellipse of size width, height inside bounding box set in Phys()
func (o *EllipseObject) Draw(win *pixelgl.Window) {
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
