package main

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

func run() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// draw quadtree
	// qt()

	r := pixel.R(10, 10, 10, 10)
	r2 := pixel.R(10, 10, 20, 20)
	log.Println(r)
	log.Println(r.Center())
	log.Println(r2.Intersect(r))
	log.Println(r2.Contains(r.Center()))
}

func main() {
	pixelgl.Run(run)
}
