package world

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

// World defines the world
type World struct {
	X, Y    float64
	Objects []Object
	Ground  Object
	gravity float64
	Atlas   *text.Atlas
}

// NewWorld returns a new world
func NewWorld(x, y float64, ground Object, gravity float64) *World {
	return &World{
		Objects: []Object{},
		X:       x,
		Y:       y,
		Ground:  ground,
		gravity: gravity,
		Atlas:   text.NewAtlas(basicfont.Face7x13, text.ASCII),
	}
}

// Update updates all the objects in the world to their next state
func (w *World) Update() {
	for _, o := range w.Objects {
		o.Update(w)
	}
}

// NextTick moves the world to the next state
func (w *World) NextTick() {

	// After update, swap the state of all objects at once
	for _, o := range w.Objects {
		o.SwapNextState()
	}
}

// AddObject adds a new object to the world
func (w *World) AddObject(o Object) {
	w.Objects = append(w.Objects, o)
}

// CheckIntersect checks if any objects in the world intersect and prints an error.
func (w *World) CheckIntersect() {
	for _, o := range w.Objects {
		for _, other := range w.Objects {
			if o.ID() == other.ID() {
				continue // skip yourself
			}
			if o.Phys().Rect.Intersect(other.Phys().Rect) != pixel.R(0, 0, 0, 0) {
				log.Printf("%#v intersects with %#v", o, other)
			}

		}
	}
}
