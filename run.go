package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"golang.org/x/image/colornames"
)

var (
	frames = 0
	second = time.Tick(time.Second)
)

const (
	// MsPerUpdate ms per game update loop, excluding rendering. This needs to be less than the time for the main update() loop
	MsPerUpdate = 4
	gravity     = -2
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
	// initial location of the object (bottom left corner)
	iX, iY float64
	// current horizontal and vertical speed of object
	vel pixel.Vec
	// currentMass of the object
	currentMass float64
	// originalMass is the original object mass
	originalMass float64
	// direction the object is facing; +1 == right; -1 == left
	dir float64
}

// update the object every frame
func (o *object) update(w *world) {
	// if above ground, fall based on mass and gravity
	if o.rect.Min.Y > w.ground.rect.Max.Y {
		// more massive objects fall faster
		o.vel.Y = w.gravity * o.currentMass
	}

	// if mass is 0, rise based on gravity
	if o.currentMass == 0 {
		o.vel.Y = -1 * w.gravity
	}

	// fall
	if o.vel.Y < 0 {
		if o.rect.Min.Y+o.vel.Y < w.ground.rect.Max.Y {
			// would fall below ground
			o.rect = o.rect.Moved(pixel.V(0, w.ground.rect.Max.Y-o.rect.Min.Y))
			o.vel.Y = 0
		} else {
			o.rect = o.rect.Moved(pixel.V(0, o.vel.Y))
		}
		return
	}

	// rise
	if o.vel.Y > 0 {
		if o.rect.Max.Y+o.vel.Y > w.Y {
			// would rise above ceiling
			o.rect = o.rect.Moved(pixel.V(0, w.Y-o.rect.Max.Y))
			o.vel.Y = 0
			o.currentMass = o.originalMass
		} else {
			o.rect = o.rect.Moved(pixel.V(0, o.vel.Y))
		}
		return
	}

	// jump back up with random probability by setting mass to 0
	if utils.RandomInt(0, 1000) < 1 {
		o.currentMass = 0 // make it float
		return
	}

	// move if on the ground
	if o.rect.Min.X <= 0 {
		o.dir = 1 // face right
	}
	if o.rect.Max.X >= w.X {
		o.dir = -1
	}
	o.rect = o.rect.Moved(pixel.V(o.dir*o.vel.X, 0))

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
	// defer utils.Elapsed("update")()
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

	objs := []*object{
		{
			color:        colornames.Red,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(1, 0),
			originalMass: 2,
			currentMass:  2,
			dir:          +1, // start facing right
			iX:           0,
			iY:           700,
		},
		{
			color:        colornames.Blue,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(1.1, 0),
			originalMass: 1,
			currentMass:  1,
			dir:          +1, // start facing right
			iX:           100,
			iY:           700,
		},
		{
			color:        colornames.Yellow,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(1.12, 0),
			originalMass: 0.5,
			currentMass:  0.5,
			dir:          +1, // start facing right
			iX:           200,
			iY:           700,
		},
	}

	for _, o := range objs {
		o.rect = pixel.R(o.iX, o.iY, o.W+o.iX, o.H+o.iY)
		w.objects = append(w.objects, o)
	}

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
		elapsed := time.Since(previous).Nanoseconds() / 1000000
		previous = time.Now()
		lag += elapsed

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
