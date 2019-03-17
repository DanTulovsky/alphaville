package world

import (
	"fmt"
	"image/color"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
)

// RectObject is a rectangular object
type RectObject struct {
	BaseObject

	// size of RectObject
	width, height float64
}

// NewRectObject return a new rectangular object
func NewRectObject(name string, color color.Color, speed, mass, width, height float64, location pixel.Rect, atlas *text.Atlas) *RectObject {

	phys := NewBaseObjectPhys(location)

	o := &RectObject{
		NewBaseObject(name, color, speed, mass, phys, atlas),
		width,
		height,
	}

	return o
}

// Draw a rectangle of size width, height inside bounding box set in Phys()
func (o *RectObject) Draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	center := o.Phys().Location().Center()
	min := pixel.V(center.X-o.width/2, center.Y-o.height/2)
	max := pixel.V(center.X+o.width/2, center.Y+o.height/2)

	// matrix to manipulate shape
	mat := pixel.IM

	o.imd.SetMatrix(mat)
	o.imd.Push(min)
	o.imd.Push(max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), o.Atlas)
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}

// NewRectObjectPhys return a new physics object
func NewRectObjectPhys(rect pixel.Rect) ObjectPhys {
	return NewBaseObjectPhys(rect)
}
