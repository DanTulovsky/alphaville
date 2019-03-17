package main

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func print(o world.Object) {
	log.Printf("Phys: %#v\n", o.Phys())
	log.Printf("Next Phys: %#v\n", o.NextPhys())
	log.Println()
}

func main() {

	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	o := world.NewRectObject(
		"one",
		colornames.Red,
		2,
		1,
		20,
		20,
		world.NewRectObjectPhys(pixel.R(0, 0, 0, 0)),
		atlas)

	// o.SetNextPhys(world.NewRectObjectPhysCopy(o.Phys().(*world.RectObjectPhys)))
	print(o)

	o.NextPhys().SetCurrentMass(10)
	o.NextPhys().SetVel(pixel.V(20, 20))

	print(o)
}
