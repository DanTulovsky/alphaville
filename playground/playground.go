package main

import (
	"log"

	"github.com/faiface/pixel/pixelgl"
)

func run() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// draw quadtree
	qt()
}

func main() {
	pixelgl.Run(run)
}
