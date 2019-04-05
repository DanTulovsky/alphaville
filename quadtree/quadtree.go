package quadtree

import (
	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Tree is a quadtree
type Tree struct {
	Bounds  pixel.Rect   // physical bounds of this node
	Level   int          // level of the node
	Objects []pixel.Rect // objects present in this node

	// 0 or 4 subnodes
	// 0 = top left, 1 = top right, 2 = bottom right, 3 = bottom left
	Nodes []*Tree // 0 or 4 subnodes
}

// NewTree returns a new quadtree
func NewTree(bounds pixel.Rect, level int) *Tree {
	return &Tree{
		Bounds:  bounds.Norm(),
		Level:   level,
		Objects: make([]pixel.Rect, 0),
		Nodes:   make([]*Tree, 0),
	}
}

// split splits the node into 4 subnodes
func (qt *Tree) split() {
	if len(qt.Nodes) == 4 {
		return
	}

	nextLevel := qt.Level
	var bounds pixel.Rect

	// top left node
	bounds = pixel.R(qt.Bounds.Min.X, qt.Bounds.Max.Y/2, qt.Bounds.Max.X/2, qt.Bounds.Max.Y)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel))

	// top right node
	bounds = pixel.R(qt.Bounds.Max.X/2, qt.Bounds.Max.Y/2, qt.Bounds.Max.X, qt.Bounds.Max.Y)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel))

	// bottom right node
	bounds = pixel.R(qt.Bounds.Max.X/2, qt.Bounds.Min.Y, qt.Bounds.Max.X, qt.Bounds.Max.Y/2)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel))

	// bottom left node
	bounds = pixel.R(qt.Bounds.Min.X, qt.Bounds.Min.Y, qt.Bounds.Max.X/2, qt.Bounds.Max.Y/2)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel))

}

// Insert inserts an object into the tree at all nodes
func (qt *Tree) Insert(r pixel.Rect) {
	if utils.Intersect(qt.Bounds, r) {
		qt.Objects = append(qt.Objects, r)
	}

	for _, n := range qt.Nodes {
		n.Insert(r)
	}
}
