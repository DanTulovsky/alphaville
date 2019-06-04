package world

import (
	behave "github.com/askft/go-behave"
	"github.com/faiface/pixel/pixelgl"
)

// Behavior is the interface for all behaviors
type Behavior interface {
	Description() string
	Draw(win *pixelgl.Window)
	Name() string
	Parent() Object
	SetParent(Object)
	Tree() *behave.BehaviorTree
	Update(*World, Object)
}
