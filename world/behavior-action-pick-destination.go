package world

import (
	"github.com/askft/go-behave/core"
)

// PickDestination picks a (random?) destination for the object to move to.
func PickDestination(params core.Params, returns core.Returns) core.Node {
	base := core.NewLeaf("PickDestination", params, returns)
	return &pickDestination{Leaf: base}
}

// succeed ...
type pickDestination struct {
	*core.Leaf
}

// Enter ...
func (a *pickDestination) Enter(ctx *core.Context) {}

// Tick ...
func (a *pickDestination) Tick(ctx *core.Context) core.Status {
	return core.StatusSuccess
}

// Leave ...
func (a *pickDestination) Leave(ctx *core.Context) {}
