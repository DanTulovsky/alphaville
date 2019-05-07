package world

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// ManualBehavior is human controlled
type ManualBehavior struct {
	DefaultBehavior
}

// NewManualBehavior return a ManualBehavior
func NewManualBehavior() *ManualBehavior {
	return &ManualBehavior{
		DefaultBehavior{
			name:        "manual_behavior",
			description: "Controlled by a human.",
		},
	}
}

// Update implements the Behavior Update method
func (b *ManualBehavior) Update(w *World, o Object) {
	phys := o.NextPhys()

	if len(phys.HaveCollisionsAt(w)) == 0 {
		b.Move(w, o, phys.CollisionBordersVector(w, phys.Vel()))
	}
}

// Move moves the object
func (b *ManualBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}

// Draw draws any artifacts of the behavior
func (b *ManualBehavior) Draw(win *pixelgl.Window) {

}
