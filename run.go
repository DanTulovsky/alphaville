package main

import (
	"fmt"
	"image/color"
	_ "image/png"
	"log"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
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
	name  string
	id    uuid.UUID
	color color.Color

	// initial speed and mass of object
	speed float64 // horizontal speed (negative means move left)
	mass  float64

	// size of the object, assuming it's a rectangle
	W, H float64

	// draws the object
	imd *imdraw.IMDraw

	// initial location of the object (bottom left corner)
	iX, iY float64

	// physics properties of the object
	phys *objectPhys
}

type objectPhys struct {

	// current horizontal and vertical speed of object
	vel pixel.Vec

	// currentMass of the object
	currentMass float64

	// this is the location of the object in the world
	rect pixel.Rect
}

// update the object every frame
func (o *object) update(w *world) {
	defer checkIntersectObject(w, o)

	// if above ground, fall based on mass and gravity
	if o.phys.rect.Min.Y > w.ground.phys.rect.Max.Y {
		// more massive objects fall faster
		o.phys.vel.Y = w.gravity * o.phys.currentMass
	}

	// if mass is 0, rise based on gravity
	if o.phys.currentMass == 0 {
		o.phys.vel.Y = -1 * w.gravity
	}

	// otherwise move
	// 	o.phys.rect = o.phys.rect.Moved(pixel.V(o.phys.vel.X, 0))

	// fall
	if o.phys.vel.Y < 0 {
		// if about to fall on another, rise back up
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if o.phys.rect.Max.X < other.phys.rect.Min.X || o.phys.rect.Min.X > other.phys.rect.Max.X {
				continue // no intersection in X axis
			}

			gap := o.phys.rect.Min.Y - other.phys.rect.Max.Y
			if !(gap >= 0 && o.phys.rect.Min.Y+o.phys.vel.Y-other.phys.rect.Max.Y <= 0) {
				// too far apart
				continue
			}

			// if about to hit another one
			switch {
			case other.phys.vel.Y < 0: // falling also
				if math.Abs(o.phys.vel.Y) > math.Abs(other.phys.vel.Y) {
					// close and falling faster than what is below
					o.phys.currentMass = 0
					o.phys.vel.Y = 0
					return

				}
			case other.phys.rect.Min.Y == w.ground.phys.rect.Max.Y:
				// close and falling on something on the ground
				o.phys.currentMass = 0
				o.phys.vel.Y = 0
				return
			case other.phys.rect.Min.Y > 0: // rising
				o.phys.currentMass = 0
				o.phys.vel.Y = 0
				return

			}
		}
		if o.phys.rect.Min.Y+o.phys.vel.Y < w.ground.phys.rect.Max.Y {
			o.phys.rect = o.phys.rect.Moved(pixel.V(0, w.ground.phys.rect.Max.Y-o.phys.rect.Min.Y))
			o.phys.vel.Y = 0
		} else {
			o.phys.rect = o.phys.rect.Moved(pixel.V(0, o.phys.vel.Y))
		}
		return
	}

	// rise
	if o.phys.vel.Y > 0 {
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if o.phys.rect.Max.X < other.phys.rect.Min.X || o.phys.rect.Min.X > other.phys.rect.Max.X {
				continue // no intersection in X axis
			}
			gap := other.phys.rect.Min.Y - o.phys.rect.Max.Y
			if gap < 0 {
				continue
			}
			// if about to hit another one
			if other.phys.rect.Min.Y-(o.phys.rect.Max.Y+o.phys.vel.Y) <= o.phys.vel.Y {
				o.phys.currentMass = o.mass
				o.phys.vel.Y = 0
				return
			}
		}

		if o.phys.rect.Max.Y+o.phys.vel.Y > w.Y {
			// would rise above ceiling
			o.phys.rect = o.phys.rect.Moved(pixel.V(0, w.Y-o.phys.rect.Max.Y))
			o.phys.vel.Y = 0
			o.phys.currentMass = o.mass

		} else {
			o.phys.rect = o.phys.rect.Moved(pixel.V(0, o.phys.vel.Y))
		}
		return
	}

	// jump back up with random probability by setting mass to 0
	// if utils.RandomInt(0, 1000) < 1 {
	// 	o.phys.currentMass = 0 // make it float
	// 	return
	// }

	// move if on the ground

	// switch directions of at the end of screen
	if o.phys.rect.Min.X <= 0 {
		o.phys.vel.X = math.Abs(o.phys.vel.X)
	}
	if o.phys.rect.Max.X >= w.X {
		o.phys.vel.X = -1 * math.Abs(o.phys.vel.X)
	}

	// if about to bump into another object, rise up
	switch {
	case o.phys.vel.X > 0: // moving right
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.phys.rect.Min.Y > o.phys.rect.Max.Y {
				continue // ignore falling objects higher than you
			}

			if o.phys.rect.Max.X <= other.phys.rect.Min.X && o.phys.rect.Max.X+o.phys.vel.X >= other.phys.rect.Min.X {
				o.phys.currentMass = 0
				return
			}
		}
	case o.phys.vel.X < 0: // moving left
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if other.phys.rect.Min.Y > o.phys.rect.Max.Y {
				continue // ignore falling objects higher than you
			}
			if o.phys.rect.Min.X >= other.phys.rect.Max.X && o.phys.rect.Min.X+o.phys.vel.X <= other.phys.rect.Max.X {
				o.phys.currentMass = 0
				return
			}
		}
	}
	// if utils.RandomInt(0, 1000) > 1 {
	// 	o.phys.currentMass = 0
	// }
	// move
	o.phys.rect = o.phys.rect.Moved(pixel.V(o.phys.vel.X, 0))

}

func (o *object) draw(win *pixelgl.Window) {
	o.imd.Clear()
	o.imd.Reset()
	o.imd.Color = o.color
	o.imd.Push(o.phys.rect.Min)
	o.imd.Push(o.phys.rect.Max)
	o.imd.Rectangle(0)
	o.imd.Draw(win)
}

func processInput() {

}

func checkIntersectObject(w *world, o *object) {
	for _, other := range w.objects {
		if o.id == other.id {
			continue // skip yourself
		}
		if o.phys.rect.Intersect(other.phys.rect) != pixel.R(0, 0, 0, 0) {
			log.Printf("%#v (%v) intersects with %#v (%v)", o.name, o.phys, other.name, other.phys)
		}
	}
}
func checkIntersect(w *world) {
	for _, o := range w.objects {
		for _, other := range w.objects {
			if o.id == other.id {
				continue // skip yourself
			}
			if o.phys.rect.Intersect(other.phys.rect) != pixel.R(0, 0, 0, 0) {
				log.Printf("%#v intersects with %#v", o, other)
			}

		}
	}

}

// update calls each object's update
func update(w *world) {
	// defer utils.Elapsed("update")()
	for _, o := range w.objects {
		o.update(w)
	}

	// checkIntersect(w)
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
			name:  "one",
			id:    uuid.New(),
			color: colornames.Red,
			imd:   imdraw.New(nil),
			W:     60,
			H:     60,
			speed: 1,
			mass:  3,
			phys:  &objectPhys{},
		},
		{
			name:  "two",
			id:    uuid.New(),
			color: colornames.Blue,
			imd:   imdraw.New(nil),
			W:     60,
			H:     60,
			speed: 1,
			mass:  2,
			phys:  &objectPhys{},
		},
		{
			name:  "three",
			id:    uuid.New(),
			color: colornames.Yellow,
			imd:   imdraw.New(nil),
			W:     60,
			H:     60,
			speed: 1.12,
			mass:  0.5,
			phys:  &objectPhys{},
		},
		{
			name:  "four",
			id:    uuid.New(),
			color: colornames.Green,
			imd:   imdraw.New(nil),
			W:     80,
			H:     80,
			speed: 0.12,
			mass:  0.1,
			phys:  &objectPhys{},
		},
		{
			name:  "five",
			id:    uuid.New(),
			color: colornames.Orange,
			imd:   imdraw.New(nil),
			W:     80,
			H:     80,
			speed: 1.12,
			mass:  4.2,
			phys:  &objectPhys{},
		},
		{
			name:  "six",
			id:    uuid.New(),
			color: colornames.Pink,
			imd:   imdraw.New(nil),
			W:     90,
			H:     20,
			speed: 1.5,
			mass:  2.2,
			phys:  &objectPhys{},
		},
		{
			name:  "seven",
			id:    uuid.New(),
			color: colornames.Brown,
			imd:   imdraw.New(nil),
			W:     30,
			H:     70,
			speed: 0.5,
			mass:  12,
			phys:  &objectPhys{},
		},
		{
			name:  "eight",
			id:    uuid.New(),
			color: colornames.Whitesmoke,
			imd:   imdraw.New(nil),
			W:     300,
			H:     70,
			speed: 0.1,
			mass:  2,
			phys:  &objectPhys{},
		},
		{
			name:  "nine",
			id:    uuid.New(),
			color: colornames.Gold,
			imd:   imdraw.New(nil),
			W:     10,
			H:     15,
			speed: 11,
			mass:  .2,
			phys:  &objectPhys{},
		},
	}

	var x float64
	for _, o := range objs {
		if o.iY == 0 {
			// place randomly, avoid intersection
			o.iY = utils.RandomFloat64(w.ground.phys.rect.Max.Y, w.Y-o.H)
		}
		if o.iX == 0 {
			// place randomly, avoid intersection
			o.iX = x
			x += o.W + 1
		}
		// set bounding rectangle based on size and location
		o.phys.rect = pixel.R(o.iX, o.iY, o.W+o.iX, o.H+o.iY)

		// set velocity vector
		o.phys.vel = pixel.V(o.speed, 0)

		// set current mass based on initial mass
		o.phys.currentMass = o.mass

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
		phys:  &objectPhys{},
	}
	ground.phys.rect = pixel.R(0, 0, 1024, 40)

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
