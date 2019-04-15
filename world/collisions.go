package world

import (
	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// HaveCollisions returns true if r1 and r2 collide
// v1 and v2 are r1 and r2 velocity vectors
func HaveCollisions(r1, r2 pixel.Rect, v1, v2 pixel.Vec) bool {

	// do the quick check first, this does not handle movements that "jump over" the object
	// only returns true if the objects would intersect during movement
	// if r1.Moved(v1).Intersect(r2.Moved(v2)) != pixel.R(0, 0, 0, 0) {
	// 	return true
	// }

	// otherwise do the longer check from
	// https://gamedev.stackexchange.com/questions/93035/whats-the-fastest-way-checking-if-two-moving-aabbs-intersect
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
	ls := pixel.L(pixel.V(0, 0), pixel.V(v.X, v.Y))

	// now count how many times ls intersects ms
	// 0 means movement did not cause collision
	// 1 means r1 ended up inside r2
	// 2 means r1 ended up on the other side of r2
	collisions := 0

	// broken due to https://github.com/faiface/pixel/issues/175
	for _, edge := range ms.Edges() {
		_, isect := edge.Intersect(ls)
		if isect {
			collisions++
		}
	}
	return collisions != 0
}
