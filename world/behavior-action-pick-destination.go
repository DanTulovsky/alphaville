package world

import (
	"github.com/askft/go-behave/core"
)

// PickDestination picks a (random?) destination for the object to move to.
func PickDestination(params core.Params, returns core.Returns) core.Node {
	base := core.NewLeaf("PickDestination", params, returns)
	return &succeed{Leaf: base}
}

// succeed ...
type succeed struct {
	*core.Leaf
}

// Enter ...
func (a *succeed) Enter(ctx *core.Context) {}

// Tick ...
func (a *succeed) Tick(ctx *core.Context) core.Status {
	return core.StatusSuccess
}

// Leave ...
func (a *succeed) Leave(ctx *core.Context) {}
