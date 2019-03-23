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

// CircleObject is a rectangular object
type CircleObject struct {
	BaseObject

	// radius of CircleObject
	radius float64
}

// NewCircleObject return a new rectangular object
func NewCircleObject(name string, color color.Color, speed, mass, radius float64) *CircleObject {

	o := &CircleObject{
		NewBaseObject(name, color, speed, mass),
		radius,
	}

	return o
}

// BoundingBox returns a Rect, rooted at the center, that covers the object
func (o *CircleObject) BoundingBox(c pixel.Vec) pixel.Rect {
	min := pixel.V(c.X-o.radius, c.Y-o.radius)
	max := pixel.V(c.X+o.radius, c.Y+o.radius)

	return pixel.R(min.X, min.Y, max.X, max.Y)
}

// Draw a rectangle of size width, height inside bounding box set in Phys()
func (o *CircleObject) Draw(win *pixelgl.Window) {
	if !o.IsSpawned() {
		return
	}

	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	center := o.Phys().Location().Center()

	o.imd.Push(center)
	o.imd.Circle(o.radius, 0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), utils.Atlas())
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}
