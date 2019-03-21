package main

import (
	"fmt"
	_ "image/png"
	"log"
	"math/rand"
	"strconv"
	"time"

	"gogs.wetsnow.com/dant/alphaville/behavior"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/populate"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

var (
	frames  = 0
	updates = 0
	second  = time.Tick(time.Second)
)

const (
	// MsPerUpdate ms per game update loop, excluding rendering. This needs to be less than the time for the main update() loop
	MsPerUpdate = 4
	gravity     = -2

	worldMaxX, worldMaxY           = 1024, 768
	visibleWinMaxX, visibleWinMaxY = 1024, 768
	groundHeight                   = 40
)

func processInput() {

}

// update calls each object's update
func update(w *world.World) {
	// defer utils.Elapsed("update")()
	w.Update()
	w.NextTick()
	w.SpawnAllNew()
}

// draw calls each object's update
func draw(w *world.World, win *pixelgl.Window) {
	w.Ground.Draw(win)

	for _, g := range w.Gates {
		g.Draw(win)
	}

	for _, f := range w.Fixtures() {
		f.Draw(win)
	}

	for _, o := range w.Objects {
		o.Draw(win)
	}
}

func run() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	groundPhys := world.NewBaseObjectPhys(pixel.R(0, 0, worldMaxX, groundHeight))
	ground := world.NewGroundObject(
		"ground", colornames.White, 0, 0, worldMaxX, groundHeight)
	ground.SetPhys(groundPhys)
	ground.SetNextPhys(ground.Phys().Copy())

	w := world.NewWorld(worldMaxX, worldMaxY, ground, gravity)

	// populate the world
	// populate.Static(world)
	populate.RandomCircles(w, 5)
	populate.RandomRectangles(w, 8)
	populate.RandomEllipses(w, 5)
	populate.AddGates(w, time.Second*5)
	populate.AddFixtures(w)

	cfg := pixelgl.WindowConfig{
		Title:  "Play!",
		Bounds: pixel.R(0, 0, visibleWinMaxX, visibleWinMaxY),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// set to false for pixel art
	win.SetSmooth(true)
	win.Clear(colornames.Black)

	previous := time.Now()
	// how far the game's clock is behind compared to the real world; in ms
	var lag int64

	// show world stats periodically
	ticker := time.NewTicker(time.Second * 5)
	go func() {
		for range ticker.C {
			w.ShowStats()
		}
	}()

	// Main loop to keep window running
	for !win.Closed() {
		elapsed := time.Since(previous).Nanoseconds() / 1000000
		previous = time.Now()
		lag += elapsed

		// user input
		processInput()

		// update the game state
		for lag >= MsPerUpdate {
			update(w)
			updates++
			lag -= MsPerUpdate
		}

		// render below here
		win.Clear(colornames.Black)
		draw(w, win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d", cfg.Title, frames))
			w.EventNotifier.Notify(
				w.NewWorldEvent(fmt.Sprintf("fps"), time.Now(),
					behavior.EventData{Key: "fps", Value: strconv.Itoa(frames)},
					behavior.EventData{Key: "ups", Value: strconv.Itoa(updates)}))
			frames = 0
			updates = 0
		default:
		}
	}
	w.End()
}

func main() {
	pixelgl.Run(run)
}
