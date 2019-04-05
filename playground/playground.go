package main

import (
	"log"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/quadtree"
)

func rotatedAroundOrigin(r pixel.Rect) pixel.Rect {
	ro := r.Moved(pixel.V(-r.Center().X, -r.Center().Y))
	return pixel.R(-ro.Min.X, -ro.Min.Y, -ro.Max.X, -ro.Max.Y).Norm()
}

func minkowskiSum(r1, r2 pixel.Rect) pixel.Rect {
	return pixel.R(r1.Min.X+r2.Min.X, r1.Min.Y+r2.Min.Y, r1.Max.X+r2.Max.X, r1.Max.Y+r2.Max.Y)
}

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// r1 := pixel.R(0, 0, 4, 4)

	// r2 := pixel.R(-2, -2, 2, 2)

	// log.Printf("r1: %v", r1)
	// v := pixel.V(r1.W()+r2.W(), r1.H()+r2.H())
	// log.Printf("resized by %v: %v", v, r1.Resized(r1.Center(), v))

	// l1 := pixel.L(pixel.V(600, 600), pixel.V(925, 150))
	// l2 := pixel.L(pixel.V(740, 255), pixel.V(925, 255))

	// l3 := pixel.L(pixel.V(600, 600), pixel.V(925, 150))
	// // l4 := pixel.L(pixel.V(740, 255), pixel.V(925, 255.24336770882053))
	// l4 := pixel.L(pixel.V(740, 255), pixel.V(925, 255.5))

	// x, isect := l1.Intersect(l2)
	// log.Printf("%v and %v intersect? %v (at: %v)", l1, l2, isect, x)

	// y, isect2 := l3.Intersect(l4)
	// log.Printf("%v and %v intersect? %v (at: %v)", l3, l4, isect2, y)

	bounds := pixel.R(0, 0, 100, 100)
	qt := quadtree.NewTree(bounds, 0, 0)

	objects := []pixel.Rect{
		// pixel.R(0, 0, 10, 10),
		// pixel.R(20, 20, 30, 30),
		// pixel.R(10, 40, 50, 50),
	}

	for _, r := range objects {
		qt.Insert(r)
	}

	log.Printf("%#v", qt)

	g := qt.ToGraph()
	log.Printf("%v", g)

	qt.Split()
	g = qt.ToGraph()
	log.Printf("%v", g)

	qt.Nodes[0].Split()
	g = qt.ToGraph()
	log.Printf("%v", g)
}
