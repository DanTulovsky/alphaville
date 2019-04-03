package world

import (
	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// HaveCollisions returns true if r1 and r2 collide
// v1 and v2 are r1 and r2 velocity vectors
func HaveCollisions(r1, r2 pixel.Rect, v1, v2 pixel.Vec) bool {

	r1 = r1.Norm()
	r2 = r2.Norm()

	// r1 moved and rotated around origin
	r1r := utils.RotatedAroundOrigin(r1)

	// r2 moved same amount as r1
	r2m := r2.Moved(pixel.V(-r1.Center().X, -r1.Center().Y))

	ms := utils.MinkowskiSum(r1r, r2m)

	// relative velocity of r1 against r2
	v := v1.Sub(v2)

	// line from origin to v
	ls := graph.Edge{A: pixel.V(0, 0), B: pixel.V(v.X, v.Y)}

	// now count how many times ls intersects ms
	// 0 means movement did not cause collision
	// 1 means r1 ended up inside r2
	// 2 means r1 ended up on the other side of r2
	collisions := 0

	for _, edge := range graph.RectEdges(ms) {
		if graph.EdgesIntersect(edge, ls) {
			collisions++
		}
	}

	return collisions != 0
}
