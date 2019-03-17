package main

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"gogs.wetsnow.com/dant/alphaville/world"
)

func print(o world.Object) {
	log.Printf("Phys: %#v\n", o.Phys())
	log.Printf("Next Phys: %#v\n", o.NextPhys())
	log.Println()
}

func run() {

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	win.Clear(colornames.Black)
	// atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

	// in degrees
	var angle = 10

	imd := imdraw.New(nil)
	// bounding box
	b := pixel.R(100, 100, 200, 200)

	// shape
	r := pixel.R(100, 100, 200, 200)
	fmt.Println(r)

	// mat := pixel.IM
	// mat = mat.Rotated(r.Center(), utils.D2R(angle))

	var x int

	for !win.Closed() {

		x = (x + angle) % 360
		// center of the bounding box
		center := b.Center()

		min := pixel.V(center.X-50, center.Y-50)
		max := pixel.V(center.X+50, center.Y+50)

		b := utils.RotateRect(r, float64(x))
		log.Printf("bounding box: %#v", b)

		win.Clear(colornames.Black)
		imd.Clear()
		imd.Reset()

		mat := pixel.IM
		mat = mat.Rotated(r.Center(), utils.D2R(float64(x)))

		imd.SetMatrix(mat)
		imd.Push(min)
		imd.Push(max)
		imd.Rectangle(0)
		imd.Draw(win)
		win.Update()

		time.Sleep(1 * time.Second)
	}
}

func main() {
	pixelgl.Run(run)
}
