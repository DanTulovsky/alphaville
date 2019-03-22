package world

import (
	"image/color"
	"math"

	"github.com/faiface/pixel"
)

// Fixture is a non-moving fixture in the world (walls, etc...)
type Fixture struct {
	RectObject
}

// NewFixture returns a new world fixture
func NewFixture(name string, color color.Color, width, height float64) *Fixture {

	f := &Fixture{
		*(NewRectObject(name, color, 0, math.MaxFloat64, width, height)),
	}

	return f
}

// Update updates the fixture for next tick
func (f *Fixture) Update(w *World) {
	// Nothing to update for fixtures right now.
}

// Place places the fixture in the world
func (f *Fixture) Place(l pixel.Vec) {
	phys := NewBaseObjectPhys(f.BoundingBox(l))
	f.SetPhys(phys)
	f.SetNextPhys(f.Phys().Copy())
}
