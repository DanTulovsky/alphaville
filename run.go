package main

import (
	"fmt"
	_ "image/png"
	"log"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	colorful "github.com/lucasb-eyer/go-colorful"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/populate"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"

	"net/http"
	_ "net/http/pprof"
)

var (
	frames  = 0
	updates = 0
	second  = time.Tick(time.Second)
	paused  = false
)

const (
	// MsPerUpdate ms per game update loop, excluding rendering. This needs to be less than the time for the main update() loop
	MsPerUpdate = 4
	gravity     = -2

	worldMaxX, worldMaxY           = 1200, 1200
	visibleWinMaxX, visibleWinMaxY = worldMaxX, worldMaxY
	groundHeight                   = 40

	maxTargets     = 3
	maxObjectSpeed = 4
)

// processMouseLeftInput handles left click
func processMouseLeftInput(w *world.World, v pixel.Vec) {
	o, err := w.ObjectClicked(v)
	if err == nil {
		log.Printf("%v", o)
	}
}

func processInput(win *pixelgl.Window, w *world.World, ctrl pixel.Vec) {

	switch {
	case win.JustPressed(pixelgl.KeySpace):
		togglePause()
	case win.JustPressed(pixelgl.KeyD):
		toggleDebug(w)
	case win.JustPressed(pixelgl.MouseButtonLeft):
		processMouseLeftInput(w, win.MousePosition())
	}

	mo := w.ManualControl
	if !mo.IsSpawned() {
		return
	}

	switch {
	case win.Pressed(pixelgl.KeyLeft):
		ctrl.X--
	case win.Pressed(pixelgl.KeyRight):
		ctrl.X++
	case win.Pressed(pixelgl.KeyUp):
		ctrl.Y++
	case win.Pressed(pixelgl.KeyDown):
		ctrl.Y--
	}

	mo.SetManualVelocity(ctrl)

}

func togglePause() {
	paused = !paused
}

func toggleDebug(w *world.World) {

}

// update calls each object's update
func update(w *world.World) {
	// defer utils.Elapsed("update")()
	w.Update()
	w.NextTick()
	w.SpawnAllNew()
}

// draw draws the world
func draw(w *world.World, win *pixelgl.Window) {
	w.Draw(win)
}

func run() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	rand.Seed(time.Now().UnixNano())

	m := pixelgl.PrimaryMonitor()
	mWidth, mHeight := m.Size()
	mWidth = mWidth - 60
	mHeight = mHeight - 60

	// start http server for pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	ground := world.NewGroundObject(
		"ground", colornames.White, 0, 0, worldMaxX, groundHeight)
	groundPhys := world.NewBaseObjectPhys(pixel.R(0, 0, worldMaxX, groundHeight), ground)
	ground.SetPhys(groundPhys)
	ground.SetNextPhys(ground.Phys().Copy())

	w := world.NewWorld(math.Min(mWidth, worldMaxX), math.Min(mHeight, worldMaxY), ground, gravity, maxObjectSpeed)

	// populate the world
	tsColors := colorful.FastHappyPalette(10)
	populate.AddTargetSeeker(w, "1", 3, tsColors[0])
	populate.AddTargetSeeker(w, "2", 4, tsColors[1])
	populate.AddTargetSeeker(w, "3", 5, tsColors[2])
	// populate.AddTargetSeeker(w, "4", 2.2, tsColors[3])

	// populate.RandomCircles(w, 2)
	populate.RandomRectangles(w, 10)
	// populate.RandomEllipses(w, 2)
	// populate.AddManualObject(w, 60, 60)
	populate.AddGates(w, time.Second*10)
	// populate.AddFixture(w)
	populate.AddFixtures(w, 8)
	// add targets AFTER fixtures
	populate.AddTarget(w, 10, maxTargets)

	cfg := pixelgl.WindowConfig{
		Title:     "Play!",
		Bounds:    pixel.R(0, 0, math.Min(mWidth, visibleWinMaxX), math.Min(mHeight, visibleWinMaxY)),
		VSync:     true,
		Resizable: false,
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
			// w.ShowStats()
		}
	}()

	// Main loop to keep window running
	for !win.Closed() {
		elapsed := time.Since(previous).Nanoseconds() / 1000000
		previous = time.Now()
		lag += elapsed

		// manual control
		ctrl := pixel.ZV
		// user input
		processInput(win, w, ctrl)

		if !paused {
			populate.AddTarget(w, 10, maxTargets)
			update(w)
			updates++
		}

		// // update the game state
		// for lag >= MsPerUpdate {
		// 	update(w)
		// 	updates++
		// 	lag -= MsPerUpdate
		// }

		// render below here
		win.Clear(colornames.Black)
		draw(w, win)
		win.Update()

		frames++
		select {
		case <-second:
			win.SetTitle(fmt.Sprintf("%s | FPS: %d | UPS: %d", cfg.Title, frames, updates))
			w.Notify(
				w.NewWorldEvent(fmt.Sprintf("fps"), time.Now(),
					observer.EventData{Key: "fps", Value: strconv.Itoa(frames)},
					observer.EventData{Key: "ups", Value: strconv.Itoa(updates)}))
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
