package graph

import (
	"fmt"
	"log"
	"math"
	"sync"

	"github.com/faiface/pixel"
	"github.com/google/uuid"
)

// Item is the data contained in the node
type Item struct {
	V pixel.Vec // vertex
}

// NewItem returns a new item
func NewItem(v pixel.Vec) Item {
	return Item{
		V: v,
	}
}

// Node is a single node in the graph
type Node struct {
	value  Item
	cost   int
	partof uuid.UUID // what node this object is part of
}

// NewNode returns a new graph node
func NewNode(i Item) *Node {
	return &Node{
		value: i,
	}
}

// Value returns the value of a node
func (n Node) Value() Item {
	return n.value
}

// Object returns the uuid of the object this node is part of
func (n Node) Object() uuid.UUID {
	return n.partof
}

// NewItemNode creates and return a new node with v as the item
func NewItemNode(u uuid.UUID, v pixel.Vec, cost int) *Node {
	return &Node{
		value: Item{
			V: v,
		},
		cost:   cost,
		partof: u,
	}

}

// String is the representation of Node
func (n *Node) String() string {
	return fmt.Sprintf("%v", n.value)
}

// Graph is the item graph (bi-directional)
type Graph struct {
	nodes []*Node
	edges map[Node][]*Node
	lock  sync.RWMutex
}

// NewGraph returns a new graph
func New() *Graph {
	nodes := make([]*Node, 0)
	edges := make(map[Node][]*Node)

	return &Graph{
		nodes: nodes,
		edges: edges,
	}
}

// Nodes returns all the nodes in the graph
func (g *Graph) Nodes() []*Node {
	return g.nodes
}

// Edges returns all the edges in the graph
func (g *Graph) Edges() map[Node][]*Node {
	return g.edges
}

// FindNode returns the node with the provided value
func (g *Graph) FindNode(v pixel.Vec) *Node {
	for _, n := range g.nodes {
		if n.value.V == v {
			return n
		}
	}
	log.Printf("cannot find: %v", v)
	log.Printf("%v", g)
	return nil
}

// AddNode adds a node to the graph
func (g *Graph) AddNode(n *Node) {
	g.lock.Lock()
	defer g.lock.Unlock()
	g.nodes = append(g.nodes, n)
}

// slow...
func edgeInList(edge *Node, edges []*Node) bool {
	for _, e := range edges {
		if edge == e {
			return true
		}
	}
	return false
}

// AddEdge adds an edge between nodes
func (g *Graph) AddEdge(n1, n2 *Node) {
	g.lock.Lock()
	defer g.lock.Unlock()

	if _, ok := g.edges[*n1]; ok {
		if !edgeInList(n2, g.edges[*n1]) {
			g.edges[*n1] = append(g.edges[*n1], n2)
		}
	} else {
		g.edges[*n1] = append(g.edges[*n1], n2)
	}

	if _, ok := g.edges[*n2]; ok {
		if !edgeInList(n1, g.edges[*n2]) {
			g.edges[*n2] = append(g.edges[*n2], n1)
		}
	} else {
		g.edges[*n2] = append(g.edges[*n2], n1)
	}
}

// String is the string representation of the graph
func (g *Graph) String() string {
	g.lock.RLock()
	defer g.lock.RUnlock()

	s := "\n"
	for i := 0; i < len(g.nodes); i++ {
		s += g.nodes[i].String() + " -> "
		near := g.edges[*g.nodes[i]]
		for j := 0; j < len(near); j++ {
			s += near[j].String() + " "
		}
		s += "\n"
	}
	return s
}

// Edge is a line segment from a to b
type Edge struct {
	A pixel.Vec
	B pixel.Vec
}

// RectEdges returns a list of edges for the given Rect
func RectEdges(r pixel.Rect) []Edge {
	return []Edge{
		{A: pixel.V(r.Min.X, r.Min.Y), B: pixel.V(r.Min.X, r.Max.Y)},
		{A: pixel.V(r.Min.X, r.Max.Y), B: pixel.V(r.Max.X, r.Max.Y)},
		{A: pixel.V(r.Max.X, r.Max.Y), B: pixel.V(r.Max.X, r.Min.Y)},
		{A: pixel.V(r.Max.X, r.Min.Y), B: pixel.V(r.Min.X, r.Min.Y)},
	}
}

// Orientation returns point orientation vs a line
// 0 --> p, q and r are colinear
// 1 --> Clockwise, below
// 2 --> Counterclockwise, above
// From: https://www.geeksforgeeks.org/check-if-two-given-line-segments-intersect/
func Orientation(p, q, r pixel.Vec) int {
	val := (q.Y-p.Y)*(r.X-q.X) - (q.X-p.X)*(r.Y-q.Y)

	switch {
	case val < 0:
		return 2
	case val > 0:
		return 1
	}

	return 0 // collinear
}

// OnSegment Given three colinear points p, q, r, the function checks if
// point q lies on line segment 'pr'
func OnSegment(p, q, r pixel.Vec) bool {
	if q.X <= math.Max(p.X, r.X) && q.X >= math.Min(p.X, r.X) &&
		q.Y <= math.Max(p.Y, r.Y) && q.Y >= math.Min(p.Y, r.Y) {
		return true
	}
	return false
}

// EdgesIntersect returns true if l1 and l2 intersect at any point
func EdgesIntersect(l1, l2 Edge) bool {

	s1 := Orientation(l1.A, l1.B, l2.A)
	s2 := Orientation(l1.A, l1.B, l2.B)
	s3 := Orientation(l2.A, l2.B, l1.A)
	s4 := Orientation(l2.A, l2.B, l1.B)

	if s1 != s2 && s3 != s4 {
		return true
	}

	// colinear
	if s1 == 0 && OnSegment(l1.A, l2.A, l1.B) {
		return true
	}
	if s2 == 0 && OnSegment(l1.A, l2.B, l1.B) {
		return true
	}
	if s3 == 0 && OnSegment(l2.A, l1.A, l2.B) {
		return true
	}
	if s4 == 0 && OnSegment(l2.A, l1.B, l2.B) {
		return true
	}

	return false
}

// PathFinder is a function that returns the path between start and dest
type PathFinder func(*Graph, pixel.Vec, pixel.Vec) ([]*Node, int, error)
