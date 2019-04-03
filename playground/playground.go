package main

import (
	"log"

	"gogs.wetsnow.com/dant/alphaville/graph"

	"github.com/faiface/pixel"
)

func rotatedAroundOrigin(r pixel.Rect) pixel.Rect {
	ro := r.Moved(pixel.V(-r.Center().X, -r.Center().Y))
	return pixel.R(-ro.Min.X, -ro.Min.Y, -ro.Max.X, -ro.Max.Y).Norm()
}

func minkowskiSum(r1, r2 pixel.Rect) pixel.Rect {
	return pixel.R(r1.Min.X+r2.Min.X, r1.Min.Y+r2.Min.Y, r1.Max.X+r2.Max.X, r1.Max.Y+r2.Max.Y)
}

func main() {

	r1 := pixel.R(0, 0, 4, 4)
	r2 := pixel.R(6, 3, 10, 7)

	r1v := pixel.V(16, 6)
	r2v := pixel.V(1, 1)

	r1 = r1.Norm()
	r2 = r2.Norm()

	log.Printf("r1: %v", r1)
	log.Printf("r2: %v", r2)

	// get A'⊕B

	// rotated r1 around the origin
	r1r := rotatedAroundOrigin(r1)
	log.Printf("rotated r1: %v", r1r)

	// move r2 relative to origin, same amount as r1
	r2m := r2.Moved(pixel.V(-r1.Center().X, -r1.Center().Y))

	ms := minkowskiSum(r1r, r2m)
	log.Printf("ms: %v", ms)

	// relative velocity
	v := r1v.Sub(r2v)
	log.Printf("relative velocity: %v", v)

	ls := graph.Edge{A: pixel.V(0, 0), B: pixel.V(v.X, v.Y)}
	log.Printf("line segment: %v", ls)

	// now count how many times ls intersects ms
	// 0 means movement did not cause collision
	// 1 means r1 ended up inside r2
	// 2 means r1 ended up on the other side of r2

	collisions := 0

	for _, edge := range graph.RectEdges(ms) {
		log.Printf("Checking edge: %v", edge)
		if graph.EdgesIntersect(edge, ls) {
			log.Printf("  found collision with %v", ls)
			collisions++
		}
	}

	log.Printf("collisions: %v", collisions)
}
