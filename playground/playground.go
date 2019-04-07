package main

import (
	"log"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/quadtree"
)

func main() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	bounds := pixel.R(0, 0, 100, 100)
	qt, err := quadtree.NewTree(bounds)
	if err != nil {
		log.Fatalf("%v", err)
	}

	objects := []pixel.Rect{
		// pixel.R(0, 0, 10, 10),
		// pixel.R(20, 20, 30, 30),
		// pixel.R(10, 40, 50, 50),
	}

	for _, r := range objects {
		log.Printf("adding: %v", r)
		// qt.Insert(r)
	}

	log.Printf("%#+v", qt)

	// // g := qt.ToGraph()
	// log.Printf("%v", g)

	// qt.Split()
	// g = qt.ToGraph()
	// log.Printf("%v", g)

	// qt.Nodes[0].Split()
	// g = qt.ToGraph()
	// log.Printf("%v", g)
}
