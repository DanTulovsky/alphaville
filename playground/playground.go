package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

func run() {

	rand.Seed(time.Now().UnixNano())
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	// draw quadtree
	// qt()

	r := pixel.R(10, 10, 10, 10)
	r2 := pixel.R(10, 10, 20, 20)
	log.Println(r)
	log.Println(r.Center())
	log.Println(r2.Intersect(r))
	log.Println(r2.Contains(r.Center()))

	log.Println(time.Second * time.Duration(utils.RandomInt(5, 10)))

	r3 := pixel.R(0, 0, 10, 20)
	r4 := pixel.R(2, 2, 4, 4)

	log.Println(r3.Intersect(r4))
	log.Println(r4.Intersect(r3))
}

func main() {
	pixelgl.Run(run)
}
