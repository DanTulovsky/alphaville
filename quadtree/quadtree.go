package quadtree

import (
	"bytes"
	"fmt"
	"image/color"
	"log"
	"math"

	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"

	"github.com/faiface/pixel"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"golang.org/x/image/colornames"
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

	root := &Node{
		bounds:  bounds.Norm(),
		color:   colornames.Gray,
		objects: objects,
		c:       make([]*Node, 4),
		size:    bounds.H(),
		level:   0,
	}

	qt := &Tree{
		root:    root,
		minSize: minSize,
	}

	qt.subdivide(qt.root)
	return qt, nil
}

func (qt *Tree) newNode(bounds pixel.Rect, parent *Node, location Quadrant) *Node {
	level := parent.level + 1

	n := &Node{
		color:    colornames.Gray,
		bounds:   bounds,
		parent:   parent,
		location: location,
		c:        make([]*Node, 4),
		size:     bounds.W(),
		level:    level,
	}

	if qt.nLevels < level {
		qt.nLevels = level
	}

	// populate the objects of this node from the parent
	for _, o := range parent.Objects() {
		if utils.Intersect(n.bounds, o) {
			n.objects = append(n.objects, o)
		}
	}

	n.color = n.CalculateColor(qt.minSize)

	// fills leaves slices
	if n.color != colornames.Gray {
		qt.leaves = append(qt.leaves, n)
	}
	return n
}

// String returns the tree as a string
func (qt *Tree) String() string {
	output := bytes.NewBufferString("")

	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "Quadtree: %v (levels: %v)\n", qt.root.bounds, qt.nLevels)
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
	// Step 1: Decomposing the colornames.Gray quadrant and updating the
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
	x1 := p.bounds.Min.X + p.bounds.W()/2
	x2 := p.bounds.Max.X

	y0 := p.bounds.Min.Y
	y1 := p.bounds.Min.Y + p.bounds.H()/2
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
	if nw.color == colornames.Gray {
		qt.subdivide(nw)
	}
	if ne.color == colornames.Gray {
		qt.subdivide(ne)
	}
	if sw.color == colornames.Gray {
		qt.subdivide(sw)
	}
	if se.color == colornames.Gray {
		qt.subdivide(se)
	}
	// p.color = colornames.Black
}

// Locate returns the Node that contains the given rect, or nil.
func (qt *Tree) Locate(pt pixel.Vec) (*Node, error) {
	// binary branching method assumes the point lies in the bounds
	cnroot := qt.root
	b := cnroot.bounds
	if !b.Contains(pt) {
		return nil, fmt.Errorf("%v does not contain %v", b, pt)
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
	x = float64(pt.X-b.Min.X) / float64(b.W())
	y = float64(pt.Y-b.Min.Y) / float64(b.H())
	k = qt.nLevels - 1
	ix := uint(x * math.Pow(2.0, float64(k)))
	iy := uint(y * math.Pow(2.0, float64(k)))

	// Now, following the branching pattern is just a matter of following, for
	// each level k in the tree, the branching indicated by the (k-1)st bit from
	// each of the x, y locational codes, it directly determines the index to
	// the appropriate child cell.  When the indexed child cell has no children,
	// the desired leaf cell has been reached and the operation is complete.
	node = cnroot
	for node.color == colornames.Gray {
		k--
		bit = 1 << k
		childIdx := (ix&bit)>>k + ((iy&bit)>>k)<<1
		node = node.c[childIdx]
	}
	return node, nil
}

// ForEachLeaf calls the given function for each leaf node of the quadtree.
//
// Successive calls to the provided function are performed in no particular
// order. The color parameter allows to loop on the leaves of a particular
// color, colornames.Black or colornames.White.
// NOTE: As by definition, colornames.Gray leaves do not exist, passing colornames.Gray to
// ForEachLeaf should return all leaves, independently of their color.
func (qt *Tree) ForEachLeaf(color color.Color, fn func(*Node)) {
	for _, n := range qt.leaves {
		if color == colornames.Gray || n.Color() == color {
			fn(n)
		}
	}
}

// ToGraph converts this tree into a graph
func (qt *Tree) ToGraph(start, target pixel.Rect) *graph.Graph {
	defer utils.Elapsed("qt converted to graph")

	g := graph.New()

	nodeNeighbors := make(map[*Node]NodeList)

	startNode, err := qt.Locate(start.Center())
	if err != nil {
		log.Fatal(err)
	}
	targetNode, err := qt.Locate(target.Center())
	if err != nil {
		log.Fatal(err)
	}

	// must set this before calculating neighbors
	startNode.SetColor(colornames.White)
	targetNode.SetColor(colornames.White)

	perNode := func(n *Node) {
		neighbors := n.Neighbors()
		nodeNeighbors[n] = neighbors
	}
	// get all the nodes + neighbors
	qt.ForEachLeaf(colornames.Gray, perNode)

	// TODO: Should be able to do this in one pass
	for node := range nodeNeighbors {
		gnode := graph.NewItemNode(uuid.New(), node.Bounds().Center(), 1)
		g.AddNode(gnode)
	}

	for node, neighbors := range nodeNeighbors {
		gnode := g.FindNode(node.Bounds().Center())
		for _, n := range neighbors {
			gneighbor := g.FindNode(n.Bounds().Center())
			g.AddEdge(gnode, gneighbor)
		}
		g.AddNode(gnode)
	}

	return g
}

// Draw draws the quadtree
// drawTree will draw the quadrants
// drawText will label the centers of quadrants
// drawObjects will draw the objects over the quadrants
func (qt *Tree) Draw(win *pixelgl.Window, drawTree, colorTree, drawText, drawObjects bool) {

	// Grab all the nodes
	rectangles := NodeList{}
	perNode := func(n *Node) {
		rectangles = append(rectangles, n)
	}
	qt.ForEachLeaf(colornames.Gray, perNode)

	imd := imdraw.New(nil)

	if colorTree {
		// rectangle itself
		for _, r := range rectangles {
			imd.Color = r.Color()
			imd.Push(r.Bounds().Min)
			imd.Push(r.Bounds().Max)
			imd.Rectangle(0)
		}
	}
	if drawTree {
		// lines around it
		for _, r := range rectangles {
			imd.Color = colornames.Red
			for _, l := range r.Bounds().Edges() {
				imd.Push(l.A)
				imd.Push(l.B)
				imd.Line(1)
			}
		}
	}

	if drawText {
		for _, r := range rectangles {
			txt := text.New(r.Bounds().Center(), utils.Atlas())
			txt.Color = colornames.Darkgray
			label := fmt.Sprintf("%v,\n%v", r.Bounds().Center().X, r.Bounds().Center().Y)
			txt.Dot.X -= txt.BoundsOf(label).W() / 2
			fmt.Fprintf(txt, "%v", label)
			txt.Draw(win, pixel.IM)
		}
	}

	if drawObjects {
		// draw the objects
		imd.Color = colornames.Yellow

		for _, r := range qt.Root().Objects() {
			imd.Push(r.Min)
			imd.Push(r.Max)
			imd.Rectangle(2)
		}
	}

	imd.Draw(win)
}
