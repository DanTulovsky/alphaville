package world

import (
	"fmt"
	"image/color"

	"gogs.wetsnow.com/dant/alphaville/utils"
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
func NewEllipseObject(name string, color color.Color, speed, mass, a, b float64, behavior Behavior) *EllipseObject {

	o := &EllipseObject{
		NewBaseObject(name, color, speed, mass, behavior),
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
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), utils.Atlas())
	txt.Color = colornames.Black

	// center the text
	txt.Dot.X -= txt.BoundsOf(o.name).W() / 2

	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}
