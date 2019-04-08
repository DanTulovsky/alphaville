package quadtree

import (
	"bytes"
	"fmt"
	"math"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Tree is a quadtree
type Tree struct {
	root   *Node
	leaves NodeList

	minSize float64 // minimum size of a side of a square
	nLevels uint    // maximum number of levels of the quadtree
}

// NewTree returns a new quadtree populated with the objects
// Inserting new objects into the tree is not currently supported
func NewTree(bounds pixel.Rect, objects []pixel.Rect, minSize float64) (*Tree, error) {

	// for now only works on squares
	if bounds.W() != bounds.H() {
		return nil, fmt.Errorf("world must be square for now, given: [%v, %v]", bounds.W(), bounds.H())
	}

	root := &Node{
		bounds:  bounds.Norm(),
		color:   Gray,
		objects: objects,
		c:       make([]*Node, 4),
		size:    bounds.H(),
	}

	qt := &Tree{
		root:    root,
		minSize: minSize,
		nLevels: 20, // arbitrary, get this based on the size of the path we need
	}

	qt.subdivide(qt.root)
	return qt, nil
}

func (qt *Tree) newNode(bounds pixel.Rect, parent *Node, location Quadrant) *Node {
	n := &Node{
		color:    Gray,
		bounds:   bounds,
		parent:   parent,
		location: location,
		c:        make([]*Node, 4),
		size:     bounds.W(),
	}

	// populate the objects of this node from the parent
	for _, o := range parent.Objects() {
		if utils.Intersect(n.bounds, o) {
			n.objects = append(n.objects, o)
		}
	}

	n.color = n.CalculateColor(qt.minSize)

	// fills leaves slices
	if n.color != Gray {
		qt.leaves = append(qt.leaves, n)
	}
	return n
}

// String returns the tree as a string
func (qt *Tree) String() string {
	output := bytes.NewBufferString("")

	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "Quadtree: %v\n", qt.root.bounds)
	fmt.Fprintf(output, "  Bounds: %v\n", qt.root.bounds)
	fmt.Fprintf(output, "  Objects: %v\n", len(qt.root.objects))
	fmt.Fprintf(output, "  Leaf Nodes: %v\n", len(qt.leaves))

	return output.String()
}

// Root returns the root node of the tree
func (qt *Tree) Root() *Node {
	return qt.root
}

func (qt *Tree) subdivide(p *Node) {
	// Step 1: Decomposing the gray quadrant and updating the
	//         parent node following the Z-order traversal.

	//     x0   x1     x2
	//  y0 .----.-------.
	//     |    |       |
	//     | NW |  NE   |
	//     |    |       |
	//  y1 '----'-------'
	//     | SW |  SE   |
	//  y2 '----'-------'
	//

	x0 := p.bounds.Min.X
	x1 := p.bounds.Min.X + p.size/2
	x2 := p.bounds.Max.X

	y0 := p.bounds.Min.Y
	y1 := p.bounds.Min.Y + p.size/2
	y2 := p.bounds.Max.Y

	// decompose current node in 4 sub-quadrants
	nw := qt.newNode(pixel.R(x0, y0, x1, y1), p, Northwest)
	ne := qt.newNode(pixel.R(x1, y0, x2, y1), p, Northeast)
	sw := qt.newNode(pixel.R(x0, y1, x1, y2), p, Southwest)
	se := qt.newNode(pixel.R(x1, y1, x2, y2), p, Southeast)

	// at creation, each sub-quadrant first inherits its parent external neighbours
	nw.cn[West] = p.cn[West]   // inherited
	nw.cn[North] = p.cn[North] // inherited
	nw.cn[East] = ne           // set for decomposition, will be updated after
	nw.cn[South] = sw          // set for decomposition, will be updated after
	ne.cn[West] = nw           // set for decomposition, will be updated after
	ne.cn[North] = p.cn[North] // inherited
	ne.cn[East] = p.cn[East]   // inherited
	ne.cn[South] = se          // set for decomposition, will be updated after
	sw.cn[West] = p.cn[West]   // inherited
	sw.cn[North] = nw          // set for decomposition, will be updated after
	sw.cn[East] = se           // set for decomposition, will be updated after
	sw.cn[South] = p.cn[South] // inherited
	se.cn[West] = sw           // set for decomposition, will be updated after
	se.cn[North] = ne          // set for decomposition, will be updated after
	se.cn[East] = p.cn[East]   // inherited
	se.cn[South] = p.cn[South] // inherited

	p.c[Northwest] = nw
	p.c[Northeast] = ne
	p.c[Southwest] = sw
	p.c[Southeast] = se

	p.updateNorthEast()
	p.updateSouthWest()

	// update all neighbours accordingly. After the decomposition
	// of a quadrant, all its neighbors in the four directions
	// must be informed of the change so that they can update
	// their own cardinal neighbors accordingly.
	p.updateNeighbours()

	// subdivide non-leaf nodes
	if nw.color == Gray {
		qt.subdivide(nw)
	}
	if ne.color == Gray {
		qt.subdivide(ne)
	}
	if sw.color == Gray {
		qt.subdivide(sw)
	}
	if se.color == Gray {
		qt.subdivide(se)
	}
	// p.color = Black
}

// Locate returns the Node that contains the given rect, or nil.
func (qt *Tree) Locate(r pixel.Rect) *Node {
	// binary branching method assumes the point lies in the bounds
	cnroot := qt.root
	b := cnroot.bounds
	if !utils.Intersect(b, r) {
		return nil
	}

	// apply affine transformations of the coordinate space, actually letting
	// the image square being defined over [0,1)²
	var (
		x, y float64
		bit  uint
		node *Node
		k    uint
	)

	// first, we multiply the position of the cell’s left corner by 2^ROOT_LEVEL
	// and then represent use product in binary form
	x = float64(r.Min.X-b.Min.X) / float64(b.W())
	y = float64(r.Min.Y-b.Min.Y) / float64(b.H())
	k = qt.nLevels - 1
	ix := uint(x * math.Pow(2.0, float64(k)))
	iy := uint(y * math.Pow(2.0, float64(k)))

	// Now, following the branching pattern is just a matter of following, for
	// each level k in the tree, the branching indicated by the (k-1)st bit from
	// each of the x, y locational codes, it directly determines the index to
	// the appropriate child cell.  When the indexed child cell has no children,
	// the desired leaf cell has been reached and the operation is complete.
	node = cnroot
	for node.color == Gray {
		k--
		bit = 1 << k
		childIdx := (ix&bit)>>k + ((iy&bit)>>k)<<1
		node = node.c[childIdx]
	}
	return node
}

// ForEachLeaf calls the given function for each leaf node of the quadtree.
//
// Successive calls to the provided function are performed in no particular
// order. The color parameter allows to loop on the leaves of a particular
// color, Black or White.
// NOTE: As by definition, Gray leaves do not exist, passing Gray to
// ForEachLeaf should return all leaves, independently of their color.
func (qt *Tree) ForEachLeaf(color Color, fn func(*Node)) {
	for _, n := range qt.leaves {
		if color == Gray || n.Color() == color {
			fn(n)
		}
	}
}

// ToGraph converts this tree into a graph
func (qt *Tree) ToGraph() *graph.Graph {
	g := graph.New()

	// Nodes of the graph are nodes of the tree with no Objects
	// or with some space not completely covered by objects
	// qt.processNode(g)

	// if !qt.isLeaf() { // tree is empty
	// 	qt.addEdges(g, nil)
	// }
	return g
}
