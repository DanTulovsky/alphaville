package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
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
	id   uuid.UUID
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
		switch {
		case o.rect.Min.Y+o.vel.Y < w.ground.rect.Max.Y:
			// would fall below ground
			o.rect = o.rect.Moved(pixel.V(0, w.ground.rect.Max.Y-o.rect.Min.Y))
			o.vel.Y = 0
		default:
			// if about to fall on another, rise back up
			for _, other := range w.objects {
				if o.id == other.id {
					continue // skip yourself
				}
				if o.rect.Max.X < other.rect.Min.X || o.rect.Min.X > other.rect.Max.X {
					continue // no intersection in X axis
				}

				gap := o.rect.Min.Y - other.rect.Max.Y
				if !(gap >= 0 && o.rect.Min.Y+o.vel.Y-other.rect.Max.Y <= 0) {
					// too far apart
					continue
				}

				// if about to hit another one
				switch {
				case other.vel.Y < 0: // falling also
					if math.Abs(o.vel.Y) > math.Abs(other.vel.Y) {
						// close and falling faster than what is below
						o.currentMass = 0
						o.vel.Y = 0
						return

					}
				case other.rect.Min.Y == w.ground.rect.Max.Y:
					// close and falling on something on the ground
					o.currentMass = 0
					o.vel.Y = 0
					return
				case other.rect.Min.Y > 0: // rising
					o.currentMass = 0
					o.vel.Y = 0
					return

				}
			}
			o.rect = o.rect.Moved(pixel.V(0, o.vel.Y))
		}
		return
	}

	// rise
	if o.vel.Y > 0 {
		switch {
		case o.rect.Max.Y+o.vel.Y > w.Y:
			// would rise above ceiling
			o.rect = o.rect.Moved(pixel.V(0, w.Y-o.rect.Max.Y))
			o.vel.Y = 0
			o.currentMass = o.originalMass
		default:
			for _, other := range w.objects {
				if o.id == other.id {
					continue // skip yourself
				}
				if o.rect.Max.X < other.rect.Min.X || o.rect.Min.X > other.rect.Max.X {
					continue // no intersection in X axis
				}
				gap := other.rect.Min.Y - o.rect.Max.Y
				if gap < 0 {
					continue
				}
				// if about to hit another one
				if other.rect.Min.Y-(o.rect.Max.Y+o.vel.Y) <= o.vel.Y {
					o.currentMass = o.originalMass
					o.vel.Y = 0
					return
				}
			}
			o.rect = o.rect.Moved(pixel.V(0, o.vel.Y))

		}
		return
	}

	// jump back up with random probability by setting mass to 0
	// if utils.RandomInt(0, 1000) < 1 {
	// 	o.currentMass = 0 // make it float
	// 	return
	// }

	// move if on the ground
	if o.rect.Min.X <= 0 {
		o.vel.X = math.Abs(o.vel.X)
	}
	if o.rect.Max.X >= w.X {
		o.vel.X = -1 * math.Abs(o.vel.X)
	}

	// if about to bump into another object, rise up
	switch {
	case o.vel.X > 0: // moving right
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.rect.Min.Y > o.rect.Max.Y {
				continue // ignore falling objects higher than you
			}

			// velocity of the other object (may be 0 if it's coming down)
			var otherVel float64
			if other.rect.Min.Y < o.rect.Max.Y {
				otherVel = 0
			} else {
				otherVel = other.vel.X
			}
			if o.rect.Max.X < other.rect.Min.X && o.rect.Max.X+o.vel.X > other.rect.Min.X+otherVel {
				o.currentMass = 0
				return
			}
		}
	case o.vel.X < 0: // moving left
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.rect.Min.Y > o.rect.Max.Y {
				continue // ignore falling objects higher than you
			}
			// velocity of the other object (may be 0 if it's coming down)
			var otherVel float64
			if other.rect.Min.Y < o.rect.Max.Y {
				otherVel = 0
			} else {
				otherVel = other.vel.X
			}
			if o.rect.Min.X > other.rect.Max.X && o.rect.Min.X+o.vel.X < other.rect.Max.X+otherVel {
				o.currentMass = 0
				return
			}
		}
	}
	// if utils.RandomInt(0, 1000) > 1 {
	// 	o.currentMass = 0
	// }
	// move
	o.rect = o.rect.Moved(pixel.V(o.vel.X, 0))

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
			name:         "one",
			id:           uuid.New(),
			color:        colornames.Red,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(1, 0),
			originalMass: 2,
			currentMass:  2,
			iX:           200,
			iY:           700,
		},
		{
			name:         "two",
			id:           uuid.New(),
			color:        colornames.Burlywood,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(2, 0),
			originalMass: 1,
			currentMass:  1,
			iX:           200,
			iY:           500,
		},
		{
			name:         "three",
			id:           uuid.New(),
			color:        colornames.Yellow,
			imd:          imdraw.New(nil),
			W:            60,
			H:            60,
			vel:          pixel.V(1.12, 0),
			originalMass: 0.5,
			currentMass:  0.5,
			iX:           200,
			iY:           700,
		},
		{
			name:         "four",
			id:           uuid.New(),
			color:        colornames.Green,
			imd:          imdraw.New(nil),
			W:            80,
			H:            80,
			vel:          pixel.V(.12, 0),
			originalMass: 0.1,
			currentMass:  0.1,
			iX:           400,
			iY:           700,
		},
		{
			name:         "four",
			id:           uuid.New(),
			color:        colornames.Green,
			imd:          imdraw.New(nil),
			W:            80,
			H:            80,
			vel:          pixel.V(.12, 0),
			originalMass: 0.2,
			currentMass:  0.2,
			iX:           500,
			iY:           700,
		},
		{
			name:         "five",
			id:           uuid.New(),
			color:        colornames.Magenta,
			imd:          imdraw.New(nil),
			W:            20,
			H:            80,
			vel:          pixel.V(1.2, 0),
			originalMass: 4.2,
			currentMass:  4.2,
			iX:           500,
			iY:           700,
		},
		{
			name:         "six",
			id:           uuid.New(),
			color:        colornames.Violet,
			imd:          imdraw.New(nil),
			W:            90,
			H:            20,
			vel:          pixel.V(1.5, 0),
			originalMass: 2.2,
			currentMass:  2.2,
			iX:           200,
			iY:           500,
		},
		{
			name:         "seven",
			id:           uuid.New(),
			color:        colornames.Teal,
			imd:          imdraw.New(nil),
			W:            30,
			H:            70,
			vel:          pixel.V(.5, 0),
			originalMass: 12,
			currentMass:  12,
			iX:           800,
			iY:           500,
		},
		{
			name:         "eight",
			id:           uuid.New(),
			color:        colornames.Whitesmoke,
			imd:          imdraw.New(nil),
			W:            300,
			H:            70,
			vel:          pixel.V(.1, 0),
			originalMass: 2,
			currentMass:  2,
			iX:           300,
			iY:           500,
		},
		{
			name:         "nine",
			id:           uuid.New(),
			color:        colornames.Tomato,
			imd:          imdraw.New(nil),
			W:            10,
			H:            15,
			vel:          pixel.V(11, 0),
			originalMass: .2,
			currentMass:  .2,
			iX:           30,
			iY:           50,
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
