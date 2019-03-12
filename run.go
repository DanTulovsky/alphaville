package main

import (
	"fmt"
	_ "image/png"
	"math/rand"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/populate"
	"gogs.wetsnow.com/dant/alphaville/world"
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

func processInput() {

}

// update calls each object's update
func update(w *world.World) {
	// defer utils.Elapsed("update")()
	for _, o := range w.Objects {
		o.Update(w)
	}

	// checkIntersect(w)
}

// draw calls each object's update
func draw(w *world.World, win *pixelgl.Window) {
	w.Ground.Draw(win)

	for _, o := range w.Objects {
		o.Draw(win)
	}
}

func run() {
	rand.Seed(time.Now().UnixNano())

	ground := world.NewObject("ground", colornames.White, 0, 0, 1024, 40, world.NewObjectPhys())
	ground.Phys.Rect = pixel.R(0, 0, 1024, 40)

	world := world.NewWorld(1024, 768, ground, gravity)

	// populate the world
	// populate.Static(world)
	populate.Random(world, 300)

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
