package quadtree

import (
	"github.com/faiface/pixel"
)

// Tree is a quadtree
type Tree struct {
	Bounds  pixel.Rect   // physical bounds of this node
	Level   int          // level of the node
	Objects []pixel.Rect // objects present in this node
	Nodes   []Tree       // 0 or 4 subnodes
}

// NewTree returns a new quadtree
func NewTree() *Tree {
	return &Tree{}
}
