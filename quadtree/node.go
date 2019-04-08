package quadtree

// from: https://github.com/arl/go-rquad/blob/master/cnnode.go

// The Western cardinal neighbor is the top-most neighbor node among the
// western neighbors, noted cn0.
//
// The Northern cardinal neighbor is the left-most neighbor node among the
// northern neighbors, noted cn1.
//
// The Eastern cardinal neighbor is the bottom-most neighbor node among the
// eastern neighbors, noted cn2.
//
// The Southern cardinal neighbor is the right-most neighbor node among the
// southern neighbors, noted cn3.

import (
	"fmt"

	"github.com/faiface/pixel"
)

// Color is the set of colors that can take a Node.
type Color byte

const (
	// Black is the color of leaf nodes
	// that are considered as obstructed.
	Black Color = 0 + iota

	// White is the color of leaf nodes
	// that are considered as free.
	White

	// Gray is the color of non-leaf nodes
	// that contain both black and white children.
	Gray
)

const colorName = "BlackWhiteGray"

var colorIndex = [...]uint8{0, 5, 10, 14}

func (i Color) String() string {
	if i >= Color(len(colorIndex)-1) {
		return fmt.Sprintf("Color(%d)", i)
	}
	return colorName[colorIndex[i]:colorIndex[i+1]]
}

// NodeList is a slice of Node pointers
type NodeList []*Node

// Node is a node in the tree
type Node struct {
	bounds  pixel.Rect   // physical bounds of this node
	objects []pixel.Rect // objects present in this node

	// 0 or 4 subnodes
	c []*Node // 0 or 4 subnodes

	parent   *Node    // parent
	color    Color    // node color
	location Quadrant // node location inside its parent

	size float64  // size of a quadrant side
	cn   [4]*Node // cardinal neighbours
}

// Objects returns the list of objects covered by this node
func (n *Node) Objects() []pixel.Rect {
	return n.objects
}

// IsEmpty returns true if the node has no objects in it
func (n *Node) IsEmpty() bool {
	return len(n.objects) == 0
}

// IsPartiallyFull returns true if the node has some space not covered by objects
// Assumes objects *cannot* overlap, an empty node returns false
func (n *Node) IsPartiallyFull() bool {

	if n.IsEmpty() {
		return false
	}

	var areaSum float64
	for _, o := range n.objects {
		areaSum += o.Area()
	}

	return areaSum < n.bounds.Area()
}

// CalculateColor returns the color the node should be
// black = complete covered by Objects or max resolution reached
// white = not covered by objects at all
// gray = partially covered by objects
func (n *Node) CalculateColor(minSize float64) Color {
	if n.IsEmpty() {
		return White
	}

	if n.IsPartiallyFull() {
		if n.size < minSize {
			return Black
		}
		return Gray
	}

	return Black
}

// Parent returns the quadtree node that is the parent of current one.
func (n *Node) Parent() *Node {
	if n.parent == nil {
		return nil
	}
	return n.parent
}

// Child returns current node child at specified quadrant.
func (n *Node) Child(q Quadrant) *Node {
	if n.c[q] == nil {
		return nil
	}
	return n.c[q]
}

// Bounds returns the bounds of the rectangular area represented by this
// quadtree node.
func (n *Node) Bounds() pixel.Rect {
	return n.bounds
}

// Color returns the node Color.
func (n *Node) Color() Color {
	return n.color
}

// Location returns the node inside its parent quadrant
func (n *Node) Location() Quadrant {
	return n.location
}

func (n *Node) updateNorthEast() {
	if n.parent == nil || n.cn[North] == nil {
		// nothing to update as this quadrant lies on the north border
		return
	}
	// step 2.2: Updating Cardinal Neighbors of NE sub-Quadrant.
	if n.cn[North] != nil {
		if n.cn[North].size < n.size {
			c0 := n.c[Northwest]
			c0.cn[North] = n.cn[North]
			// to update C1, we perform a west-east traversal
			// recording the cumulative size of traversed nodes
			cur := c0.cn[North]
			cumsize := cur.size
			for cumsize < c0.size {
				cur = cur.cn[East]
				cumsize += cur.size
			}
			n.c[Northeast].cn[North] = cur
		}
	}
}

func (n *Node) updateSouthWest() {
	if n.parent == nil || n.cn[West] == nil {
		// nothing to update as this quadrant lies on the west border
		return
	}
	// step 2.1: Updating Cardinal Neighbors of SW sub-Quadrant.
	if n.cn[North] != nil {
		if n.cn[North].size < n.size {
			c0 := n.c[Northwest]
			c0.cn[North] = n.cn[North]
			// to update C2, we perform a north-south traversal
			// recording the cumulative size of traversed nodes
			cur := c0.cn[West]
			cumsize := cur.size
			for cumsize < c0.size {
				cur = cur.cn[South]
				cumsize += cur.size
			}
			n.c[Southwest].cn[West] = cur
		}
	}
}

// updateNeighbours updates all neighbours according to the current
// decomposition.
func (n *Node) updateNeighbours() {
	// On each direction, a full traversal of the neighbors
	// should be performed.  In every quadrant where a reference
	// to the parent quadrant is stored as the Cardinal Neighbor,
	// it should be replaced by one of its children created after
	// the decomposition

	if n.cn[West] != nil {
		n.forEachNeighbourInDirection(West, func(qn *Node) {
			western := qn
			if western.cn[East] == n {
				if western.bounds.Max.Y > n.c[Southwest].bounds.Min.Y {
					// choose SW
					western.cn[East] = n.c[Southwest]
				} else {
					// choose NW
					western.cn[East] = n.c[Northwest]
				}
				if western.cn[East].bounds.Min.Y == western.bounds.Min.Y {
					western.cn[East].cn[West] = western
				}
			}
		})
	}

	if n.cn[North] != nil {
		n.forEachNeighbourInDirection(North, func(qn *Node) {
			northern := qn
			if northern.cn[South] == n {
				if northern.bounds.Max.X > n.c[Northeast].bounds.Min.X {
					// choose NE
					northern.cn[South] = n.c[Northeast]
				} else {
					// choose NW
					northern.cn[South] = n.c[Northwest]
				}
				if northern.cn[South].bounds.Min.X == northern.bounds.Min.X {
					northern.cn[South].cn[North] = northern
				}
			}
		})
	}

	if n.cn[East] != nil {
		if n.cn[East] != nil && n.cn[East].cn[West] == n {
			// To update the eastern CN of a quadrant Q that is being
			// decomposed: Q.CN2.CN0=Q.Ch[NE]
			n.cn[East].cn[West] = n.c[Northeast]
		}
	}

	if n.cn[South] != nil {
		// To update the southern CN of a quadrant Q that is being
		// decomposed: Q.CN3.CN1=Q.Ch[SE]
		// TODO: there seems to be a typo in the paper.
		// should have read this instead: Q.CN3.CN1=Q.Ch[SW]
		if n.cn[South] != nil && n.cn[South].cn[North] == n {
			n.cn[South].cn[North] = n.c[Southwest]
		}
	}
}

// forEachNeighbourInDirection calls fn on every neighbour of the current node in the given
// direction.
func (n *Node) forEachNeighbourInDirection(dir Side, fn func(*Node)) {
	// start from the cardinal neighbour on the given direction
	N := n.cn[dir]
	if N == nil {
		return
	}
	fn(N)
	if N.size >= n.size {
		return
	}

	traversal := traversal(dir)
	opposite := opposite(dir)
	// perform cardinal neighbour traversal
	for {
		N = N.cn[traversal]
		if N != nil && N.cn[opposite] == n {
			fn(N)
		} else {
			return
		}
	}
}

// forEachNeighbour calls the given function for each neighbour of current
// node.
func (n *Node) forEachNeighbour(fn func(*Node)) {
	n.forEachNeighbourInDirection(West, fn)
	n.forEachNeighbourInDirection(North, fn)
	n.forEachNeighbourInDirection(East, fn)
	n.forEachNeighbourInDirection(South, fn)
}

// Neighbors returns the neighbors of the node
func (n *Node) Neighbors() NodeList {

	neighbors := []*Node{}

	forNeighbor := func(n *Node) {
		neighbors = append(neighbors, n)
	}

	ForEachNeighbour(n, forNeighbor)

	return neighbors
}
