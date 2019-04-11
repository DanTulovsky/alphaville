package world

import (
	"bytes"
	"fmt"
	"html/template"
	"image/color"
	"log"

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

// String returns the object as string
func (o *RectObject) String() string {
	buf := bytes.NewBufferString("")

	tmpl, err := template.New("object").Parse(
		`
----------------------------------------
Name: {{.Name}}
Speed: {{.Speed}}
Mass: {{.Mass}}
`)

	if err != nil {
		log.Fatalf("object conversion error: %v", err)
	}
	err = tmpl.Execute(buf, o)
	if err != nil {
		log.Fatalf("object conversion error: %v", err)
	}

	fmt.Fprintf(buf, "  %v", o.Phys())
	fmt.Fprintf(buf, "  %v", o.Behavior())
	fmt.Fprintf(buf, "----------------------------------------")
	return buf.String()
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
func NewRectObject(name string, color color.Color, speed, mass, width, height float64, behavior Behavior) *RectObject {

	o := &RectObject{
		NewBaseObject(name, color, speed, mass),
		width,
		height,
	}
	if behavior == nil {
		behavior = NewDefaultBehavior()
	}
	o.behavior = behavior
	o.behavior.SetParent(o)

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

	// center the text
	label := []string{o.name}

	switch b := o.behavior.(type) {
	case *TargetSeekerBehavior:
		label = append(label, fmt.Sprintf("%v, %v", b.TargetsCaught(), b.TurnsBlocked()))
	}

	for _, l := range label {
		txt.Dot.X -= txt.BoundsOf(l).W() / 2
		fmt.Fprintf(txt, "%v\n", l)
	}

	txt.Draw(win, pixel.IM)
}
