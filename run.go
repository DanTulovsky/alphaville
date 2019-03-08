package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"golang.org/x/image/colornames"
)

var (
	frames = 0
	second = time.Tick(time.Second)
)

const (
	// MsPerUpdate ms per game update loop, excluding rendering
	MsPerUpdate = 16
	gravity     = -0.02
)

type world struct {
	X, Y    float64
	objects []*object
	ground  *object
	gravity float64
}

// object is an object in the world
type object struct {
	name string
	// this is the location of the object in the world
	rect  pixel.Rect
	color color.Color
	imd   *imdraw.IMDraw
	// size of the object
	W, H float64
	vel  pixel.Vec
}

// update the object every frame
func (o *object) update(w *world) {
	if o.rect.Min.Y > w.ground.rect.Max.Y {
		o.vel.Y += w.gravity
	}

	// fall
	if o.vel.Y < 0 {
		if o.rect.Min.Y+o.vel.Y < w.ground.rect.Max.Y {
			// would fall below ground
			o.rect = o.rect.Moved(pixel.V(0, w.ground.rect.Max.Y-o.rect.Min.Y))
		} else {
			o.rect = o.rect.Moved(pixel.V(0, w.gravity))
		}
	}
	o.vel.Y = 0

}

func (o *object) draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color
	o.imd.Push(o.rect.Min)
	o.imd.Push(o.rect.Max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)
}

func processInput() {

}

// update calls each object's update
func update(w *world) {
	for _, o := range w.objects {
		o.update(w)
	}
}

// draw calls each object's update
func draw(w *world, win *pixelgl.Window) {
	w.ground.draw(win)

	for _, o := range w.objects {
		o.draw(win)
	}
}

// populate puts some objects in the world
func populate(w *world) {

	o := &object{
		color: colornames.Red,
		imd:   imdraw.New(nil),
		W:     60,
		H:     60,
	}
	o.rect = pixel.R(0, 700, o.W, o.H+700)

	w.objects = append(w.objects, o)
}

func run() {
	ground := &object{
		name:  "ground",
		color: colornames.White,
		imd:   imdraw.New(nil),
		W:     1024,
		H:     40,
	}
	ground.rect = pixel.R(0, 0, 1024, 40)

	world := &world{
		objects: []*object{},
		X:       1024,
		Y:       768,
		ground:  ground,
		gravity: gravity, // how fast objects fall
	}

	cfg := pixelgl.WindowConfig{
		Title:  "Play!",
		Bounds: pixel.R(0, 0, world.X, world.Y),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// set to false for pixel art
	win.SetSmooth(false)
	win.Clear(colornames.Black)

	populate(world)

	previous := time.Now()
	// how far the game's clock is behind compared to the real world; in ms
	var lag int64

	// Main loop to keep window running
	for !win.Closed() {
		elapsed := time.Since(previous).Nanoseconds()
		lag += elapsed / 1000000

		// user input
		processInput()

		// update the game state
		for lag >= MsPerUpdate {
			update(world)
			lag -= MsPerUpdate
		}

		// render below here
		win.Clear(colornames.Black)
		draw(world, win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			frames = 0
		default:
		}
	}
}

func main() {
	pixelgl.Run(run)
}
