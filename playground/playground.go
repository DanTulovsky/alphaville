package main

import (
	"fmt"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

func main() {
	f := world.NewFixture("wall1", colornames.Yellow, 10, 200)
	l := pixel.V(40, 100)
	f.Place(l)

	fmt.Printf("%#+v", f.Phys())
}
