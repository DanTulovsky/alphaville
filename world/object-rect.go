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

// RectObject is a rectangular object
type RectObject struct {
	BaseObject

	// size of RectObject
	width, height float64
}

// NewGroundObject return a new ground object
func NewGroundObject(name string, color color.Color, speed, mass, width, height float64) *RectObject {

	o := &RectObject{
		NewBaseObject(name, color, speed, mass),
		width,
		height,
	}

	return o
}

// NewRectObject return a new rectangular object
func NewRectObject(name string, color color.Color, speed, mass, width, height float64) *RectObject {

	o := &RectObject{
		NewBaseObject(name, color, speed, mass),
		width,
		height,
	}

	return o
}

// BoundingBox returns a Rect, rooted at the center, that covers the object
func (o *RectObject) BoundingBox(c pixel.Vec) pixel.Rect {
	min := pixel.V(c.X-o.width/2, c.Y-o.height/2)
	max := pixel.V(c.X+o.width/2, c.Y+o.height/2)

	return pixel.R(min.X, min.Y, max.X, max.Y)

}

// Draw a rectangle of size width, height inside bounding box set in Phys()
func (o *RectObject) Draw(win *pixelgl.Window) {

	if !o.IsSpawned() {
		return
	}

	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color

	box := o.BoundingBox(o.Phys().Location().Center())

	// matrix to manipulate shape
	mat := pixel.IM

	o.imd.SetMatrix(mat)
	o.imd.Push(box.Min)
	o.imd.Push(box.Max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), utils.Atlas())
	txt.Color = colornames.Black
	fmt.Fprintf(txt, "%v", o.name)
	txt.Draw(win, pixel.IM)
}
