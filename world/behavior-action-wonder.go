package world

import (
	"github.com/askft/go-behave/core"
	"github.com/faiface/pixel"
)

// Wonder picks a (random?) destination for the object to move to.
func Wonder(params core.Params, returns core.Returns) core.Node {
	base := core.NewLeaf("PickDestination", params, returns)
	return &wonder{Leaf: base}
}

// wonder ...
type wonder struct {
	*core.Leaf
	o Object
	b *WondererBehavior
	w *World
}

// Enter ...
func (a *wonder) Enter(ctx *core.Context) {

	a.o = ctx.Owner.(Object)
	a.b = a.o.Behavior().(*WondererBehavior)
	a.w = ctx.Data.(*World)
}

// Tick ...
func (a *wonder) Tick(ctx *core.Context) core.Status {

	if a.b.ChangeVerticalDirection(a.w, a.o) {
		return core.StatusRunning
	}

	if a.b.HandleCollisions(a.w, a.o) {
		return core.StatusRunning
	}

	phys := a.o.NextPhys()
	a.b.Move(a.w, a.o, pixel.V(phys.Vel().X, phys.Vel().Y))

	return core.StatusRunning
}

// Leave ...
func (a *wonder) Leave(ctx *core.Context) {

}
