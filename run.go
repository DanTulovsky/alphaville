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
	camPos       = pixel.ZV // camera position
	camSpeed     = 500.0    // camera speed in pixels per second
	camZoom      = 1.0      // 1 = no zoom
	camZoomSpeed = 1.2

	frames = 0
	second = time.Tick(time.Second)
)

const (
	// MsPerUpdate ms per game update loop, excluding rendering
	MsPerUpdate = 16
)

func processInput() {

}

func update() {

}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "Play!",
		Bounds: pixel.R(0, 0, 1024, 768),
		VSync:  true,
	}

	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	imd := imdraw.New(nil)
	imd.Color = color.Black
	imd.EndShape = imdraw.RoundEndShape
	imd.Push(pixel.V(10, 400))
	imd.Push(pixel.V(1014, 400))
	imd.Line(10)

	// set to false for pixel art
	win.SetSmooth(false)
	win.Clear(colornames.Aqua)

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
			update()
			lag -= MsPerUpdate
		}

		// render below here
		win.Clear(colornames.Green)
		imd.Draw(win)
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
