package quadtree

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// 0 = top left, 1 = top right, 2 = bottom right, 3 = bottom left
type location int

const (
	topLeft location = iota
	topRight
	bottomRight
	bottomLeft
)

// Tree is a quadtree
type Tree struct {
	Bounds  pixel.Rect   // physical bounds of this node
	Level   int          // level of the node
	Objects []pixel.Rect // objects present in this node

	// 0 or 4 subnodes
	Nodes []*Tree // 0 or 4 subnodes

	Location location
}

// NewTree returns a new quadtree
func NewTree(bounds pixel.Rect, level int, loc location) *Tree {
	return &Tree{
		Bounds:   bounds.Norm(),
		Level:    level,
		Location: loc,
		Objects:  make([]pixel.Rect, 0),
		Nodes:    make([]*Tree, 0),
	}
}

// Split splits the node into 4 subnodes
func (qt *Tree) Split() {
	if len(qt.Nodes) == 4 {
		return
	}

	nextLevel := qt.Level
	var bounds pixel.Rect

	// top left node
	bounds = pixel.R(qt.Bounds.Min.X, qt.Bounds.Max.Y/2, qt.Bounds.Max.X/2, qt.Bounds.Max.Y)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel, topLeft))

	// top right node
	bounds = pixel.R(qt.Bounds.Max.X/2, qt.Bounds.Max.Y/2, qt.Bounds.Max.X, qt.Bounds.Max.Y)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel, topRight))

	// bottom right node
	bounds = pixel.R(qt.Bounds.Max.X/2, qt.Bounds.Min.Y, qt.Bounds.Max.X, qt.Bounds.Max.Y/2)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel, bottomRight))

	// bottom left node
	bounds = pixel.R(qt.Bounds.Min.X, qt.Bounds.Min.Y, qt.Bounds.Max.X/2, qt.Bounds.Max.Y/2)
	qt.Nodes = append(qt.Nodes, NewTree(bounds, nextLevel, bottomLeft))

	// populate the objects into subnodes
	for _, o := range qt.Objects {
		for _, n := range qt.Nodes {
			if utils.Intersect(o, n.Bounds) {
				n.Insert(o)
			}
		}
	}
}

// Insert inserts an object into the tree at all nodes
func (qt *Tree) Insert(r pixel.Rect) {

	if utils.Intersect(qt.Bounds, r) {
		qt.Objects = append(qt.Objects, r)
	}

	for _, n := range qt.Nodes {
		n.Insert(r.Norm())
	}
}

// IsEmpty returns true if the node has no objects in it
func (qt *Tree) IsEmpty() bool {
	return len(qt.Objects) == 0
}

// IsPartiallyFull returns true if the node has some space not covered by objects
// Assumes objects *cannot* overlap, an empty tree returns false
func (qt *Tree) IsPartiallyFull() bool {

	if len(qt.Objects) == 0 {
		return false
	}

	var areaSum float64
	for _, o := range qt.Objects {
		areaSum += o.Area()
	}

	return areaSum < qt.Bounds.Area()
}

// isLeaf returns true if this is a leaf node
func (qt *Tree) isLeaf() bool {
	return len(qt.Nodes) == 0
}

// processNode adds the tree node to the graph if it's empty or partially covered by objects
func (qt *Tree) processNode(g *graph.Graph) {

	if qt.isLeaf() && (qt.IsEmpty() || qt.IsPartiallyFull()) {
		// the graph node is the center point of the bounds rectangle of the tree node
		node := graph.NewItemNode(uuid.New(), qt.Bounds.Center(), 1)
		g.AddNode(node)
	}

	for _, n := range qt.Nodes {
		n.processNode(g)
	}
}

// addEdges adds the edges to the graph of nodes
func (qt *Tree) addEdges(g *graph.Graph, pathFrom *Tree) {

	// process children at the same level
	var self *Tree

	if pathFrom != nil {
		switch pathFrom.Location {
		case topLeft:
			switch qt.Location {
			case topRight:
				if qt.Nodes[0].isLeaf() {
					g.AddEdge(g.FindNode(qt.Nodes[0].Bounds.Center()), g.FindNode(pathFrom.Bounds.Center()))
				} else {
					qt.Nodes[0].addEdges(g, pathFrom)
				}
				if qt.Nodes[3].isLeaf() {
					g.AddEdge(g.FindNode(qt.Nodes[3].Bounds.Center()), g.FindNode(pathFrom.Bounds.Center()))
				} else {
					qt.Nodes[3].addEdges(g, pathFrom)
				}
			case bottomLeft:
				if qt.Nodes[0].isLeaf() {
					g.AddEdge(g.FindNode(qt.Nodes[0].Bounds.Center()), g.FindNode(pathFrom.Bounds.Center()))
				} else {
					qt.Nodes[0].addEdges(g, pathFrom)
				}
				if qt.Nodes[1].isLeaf() {
					g.AddEdge(g.FindNode(qt.Nodes[1].Bounds.Center()), g.FindNode(pathFrom.Bounds.Center()))
				} else {
					qt.Nodes[1].addEdges(g, pathFrom)
				}
			}
		}
	}

	if qt.isLeaf() {
		return
	}

	// 0 is top left
	self = qt.Nodes[0]
	log.Printf("> %v", self.Bounds.Center())
	if qt.Nodes[1].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[1].Bounds.Center()))
	} else {
		qt.Nodes[1].addEdges(g, self)
	}
	if qt.Nodes[3].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[3].Bounds.Center()))
	}

	// 1 is top right
	self = qt.Nodes[1]
	if qt.Nodes[0].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[0].Bounds.Center()))
	}
	if qt.Nodes[2].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[2].Bounds.Center()))
	}

	// 2 is bottom right
	self = qt.Nodes[2]
	if qt.Nodes[1].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[1].Bounds.Center()))
	}
	if qt.Nodes[3].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[3].Bounds.Center()))
	}

	// 3 is bottom left
	self = qt.Nodes[3]
	if qt.Nodes[0].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[0].Bounds.Center()))
	}
	if qt.Nodes[2].isLeaf() {
		g.AddEdge(g.FindNode(self.Bounds.Center()), g.FindNode(qt.Nodes[2].Bounds.Center()))
	}
}

// ToGraph converts this tree into a graph
func (qt *Tree) ToGraph() *graph.Graph {
	g := graph.New()

	// Nodes of the graph are nodes of the tree with no Objects
	// or with some space not completely covered by objects
	qt.processNode(g)

	if !qt.isLeaf() { // tree is empty
		qt.addEdges(g, nil)
	}
	return g
}
