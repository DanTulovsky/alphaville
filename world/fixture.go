package world

import (
	"image/color"
	"math"
)

// Fixture is a non-moving fixture in the world (walls, etc...)
type Fixture struct {
	BaseObject

	// size of RectObject
	width, height float64
}

// NewFixture returns a new world fixture
func NewFixture(name string, color color.Color, width, height float64) *Fixture {

	f := &Fixture{
		// 0 speed, max mass
		NewBaseObject(name, color, 0, math.MaxFloat64, fixtureType),
		width,
		height,
	}
	return f
}
