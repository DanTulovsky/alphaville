package populate

import (
	"fmt"
	"image/color"
	"log"
	"math"
	"time"

	"github.com/faiface/pixel"
	colorful "github.com/lucasb-eyer/go-colorful"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

// RandomEllipses populates the world with N random objects
func RandomEllipses(w *world.World, n int) {

	var minRadius, maxRadius, minMass, maxMass, minSpeed, maxSpeed float64

	minRadius = w.MinObjectSide / 2
	maxRadius = 60
	minMass, maxMass = 1, 10
	minSpeed, maxSpeed = 0.1, w.MaxObjectSpeed

	for i := 0; i < n; i++ {

		a := utils.RandomFloat64(minRadius, maxRadius+1)
		b := utils.RandomFloat64(minRadius, maxRadius+1)

		o := world.NewEllipseObject(
			fmt.Sprintf("%v", i),
			colorful.FastWarmColor(),
			utils.RandomFloat64(minSpeed, maxSpeed),  // speed
			utils.RandomFloat64(minMass, maxMass)/10, // mass
			a,   // x radius
			b,   // y radius
			nil, // default behavior
		)

		if err := w.AddObject(o); err != nil {
			log.Fatalf("cannot add object: %v", err)
		}
	}
}

// RandomCircles populates the world with N random objects
func RandomCircles(w *world.World, n int) {

	var minRadius, maxRadius, minMass, maxMass, minSpeed, maxSpeed float64

	minRadius = w.MinObjectSide / 2
	maxRadius = 60
	minMass, maxMass = 1, 10
	minSpeed, maxSpeed = 0.1, w.MaxObjectSpeed

	for i := 0; i < n; i++ {
		radius := utils.RandomFloat64(minRadius, maxRadius+1)

		o := world.NewCircleObject(
			fmt.Sprintf("%v", i),
			colorful.FastWarmColor(),
			utils.RandomFloat64(minSpeed, maxSpeed),  // speed
			utils.RandomFloat64(minMass, maxMass)/10, // mass
			radius, // radius
			nil,    // default behavior
		)

		if err := w.AddObject(o); err != nil {
			log.Fatalf("cannot add object: %v", err)
		}
	}
}

// RandomRectangles populates the world with N random rectangular objects
func RandomRectangles(w *world.World, n int) {

	var minWidth, maxWidth, minHeight, maxHeight, minMass, maxMass, minSpeed, maxSpeed float64

	minWidth, maxWidth = w.MinObjectSide, 81
	minHeight, maxHeight = w.MinObjectSide, 81
	minMass, maxMass = 6, 10
	minSpeed, maxSpeed = 0.1, w.MaxObjectSpeed

	for i := 0; i < n; i++ {

		width := utils.RandomFloat64(minWidth, maxWidth+1)
		height := utils.RandomFloat64(minHeight, maxHeight)

		o := world.NewRectObject(
			fmt.Sprintf("%v", i),
			colorful.FastWarmColor(),
			utils.RandomFloat64(minSpeed, maxSpeed),  // speed
			utils.RandomFloat64(minMass, maxMass)/10, // mass
			width,  // width
			height, // height
			nil,    // default behavior
		)

		if err := w.AddObject(o); err != nil {
			log.Fatalf("cannot add object: %v", err)
		}
	}
}

// AddTargetSeeker adds an object that seeks a target
func AddTargetSeeker(w *world.World, name string, speed float64, c color.Color) {

	var minWidth, maxWidth, minHeight, maxHeight, minMass, maxMass float64

	minWidth, maxWidth = w.MinObjectSide+20, w.MinObjectSide+21
	minHeight, maxHeight = w.MinObjectSide+20, w.MinObjectSide+21
	minMass, maxMass = 6, 10
	// minSpeed, maxSpeed = 2, w.MaxObjectSpeed

	width := utils.RandomFloat64(minWidth, maxWidth)
	height := utils.RandomFloat64(minHeight, maxHeight)

	if c == nil {
		c = colorful.FastWarmColor()
	}

	// path finder algorithm
	finder := graph.DijkstraPath

	o := world.NewRectObject(
		fmt.Sprintf("ts-%v", name),
		c,
		speed,
		utils.RandomFloat64(minMass, maxMass)/10, // mass
		width,  // width
		height, // height
		world.NewTargetSeekerBehavior(graph.PathFinder(finder)),
	)

	if err := w.AddObject(o); err != nil {
		log.Fatalf("cannot add object: %v", err)
	}
}

// AddManualObject adds a manually controlled object to the world
func AddManualObject(w *world.World, width, height float64) {

	behavior := world.NewManualBehavior()

	o := world.NewRectObject(
		"manual",
		colornames.Red,
		3,      // speed
		1,      // mass
		width,  // width
		height, // height
		behavior,
	)

	if err := w.AddObject(o); err != nil {
		log.Fatalf("cannot add object: %v", err)
	}
	w.ManualControl = o
}

// AddGates adds gates to the world
func AddGates(w *world.World) {

	var filterManualOnly world.GateFilter = func(o world.Object) bool {
		switch o.Behavior().(type) {
		case *world.ManualBehavior:
			return true
		}
		return false
	}

	// var filterTargetSeekerOnly world.GateFilter = func(o world.Object) bool {
	// 	switch o.Behavior().(type) {
	// 	case *world.TargetSeekerBehavior:
	// 		return true
	// 	}
	// 	return false
	// }

	type gate struct {
		name     string
		location pixel.Vec
		status   world.GateStatus
		coolDown time.Duration
		radius   float64
		filters  []world.GateFilter
	}

	// add spawn gate
	gates := []gate{
		{
			name:     "One",
			location: pixel.V(600, 600),
			status:   world.GateOpen,
			coolDown: time.Second * time.Duration(utils.RandomInt(2, 10)),
			radius:   20,
			// filters:  []world.GateFilter{filterTargetSeekerOnly},
		},
		{
			name:     "Two",
			location: pixel.V(200, 600),
			status:   world.GateOpen,
			coolDown: time.Second * time.Duration(utils.RandomInt(2, 10)),
			radius:   25,
		},
		{
			name:     "manual only",
			location: pixel.V(400, 600),
			status:   world.GateOpen,
			coolDown: time.Second * time.Duration(utils.RandomInt(2, 10)),
			radius:   25,
			filters:  []world.GateFilter{filterManualOnly},
		},
	}

	for _, g := range gates {
		gate := world.NewGate(g.name, g.location, g.status, g.coolDown, g.radius, g.filters...)

		if err := w.AddGate(gate); err != nil {
			log.Fatalf("error adding gate: %v", err)
		}
	}
}

// AddTarget adds targets to the world
func AddTarget(w *world.World, radius float64, maxTargets int) error {

	if len(w.Targets()) >= maxTargets {
		return fmt.Errorf("already too many targets %v of %v", len(w.Targets()), maxTargets)
	}

	var valid bool
	var t world.Target

	// don't let targets appear inside fixtures
	for !valid {
		l := pixel.V(
			// TODO fix these!
			utils.RandomFloat64(55, w.X-65),
			utils.RandomFloat64(w.Ground.Phys().Location().Max.Y+65, w.Y-65))

		t = world.NewSimpleTarget("one", l, radius, "desc")
		valid = true
		for _, f := range w.Fixtures() {
			// log.Printf("checking fixture: %v (target: %v)", f.Phys().Location(), t.Circle().Resized(20))
			// for now assume seekers are always MinObjectSize, MinObjectSize rectangles, don't let targets end up inside
			// augmented area of fixtures, do this by resizing the circle by half the width of the rect
			if f.Phys().Location().IntersectCircle(t.Circle().Resized(w.MinObjectSide+4)) != pixel.ZV {
				valid = false
			}
		}

		// check edges of the world
		if w.Ground.Phys().Location().IntersectCircle(t.Circle().Resized(w.MinObjectSide+4)) != pixel.ZV {
			valid = false
		}
	}
	if err := w.AddTarget(t); err != nil {
		return err
	}

	return nil
}

// AddFixture adds one specific fixture to the world
func AddFixture(w *world.World) error {

	var width float64 = 144
	var height float64 = 64
	f := world.NewFixture("one", colorful.FastWarmColor(), width, height)
	f.Place(pixel.V(761, 171))
	if err := w.AddFixture(f); err != nil {
		return err
	}
	return nil
}

// AddFixtures add fixtures to the world.
func AddFixtures(w *world.World, numFixtures int) {

	var minWidth float64 = w.MinObjectSide
	var maxWidth float64 = w.MinObjectSide + 30
	var minHeight float64 = w.MinObjectSide
	var maxHeight float64 = w.Y - 100

	fcolors := colorful.FastWarmPalette(numFixtures)

	for x := 0; x < numFixtures; x++ {

		intersect := true
		var f *world.Fixture

		// These can appear closer than target seeker size, and confuse the graphgenerating algorithm
		// should be fixed  by switching to trapezoid map instead
		for intersect {
			intersect = false
			width := math.Floor(utils.RandomFloat64(minWidth, maxWidth))
			height := math.Floor(utils.RandomFloat64(minHeight, maxHeight))
			lX := utils.RandomFloat64(width, w.X-width)
			lY := utils.RandomFloat64(height, w.Y-height)

			f = world.NewFixture(fmt.Sprintf("block-%v", x), fcolors[x], width, height)
			f.Place(pixel.V(lX, lY))

			for _, other := range w.Fixtures() {
				if f.Phys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
					intersect = true
				}
			}

			for _, g := range w.Gates {
				if f.Phys().Location().Intersect(g.BoundingBox(g.Location)) != pixel.R(0, 0, 0, 0) {
					intersect = true
				}
			}

		}
		if err := w.AddFixture(f); err != nil {
			log.Fatal(err)
		}
	}

	// f := world.NewFixture("block1", colornames.Green, width, height)
	// f.Place(pixel.V(400, w.Ground.Phys().Location().Max.Y+160))
	// w.AddFixture(f)

	// f = world.NewFixture("block2", colornames.Green, width, height)
	// f.Place(pixel.V(50, w.Ground.Phys().Location().Max.Y+100))
	// w.AddFixture(f)

	// f = world.NewFixture("block3", colornames.Green, width, height)
	// f.Place(pixel.V(600, 400))
	// w.AddFixture(f)

	// f = world.NewFixture("block4", colornames.Green, width, height)
	// f.Place(pixel.V(60, 400))
	// w.AddFixture(f)

	// width = 20
	// height = 400
	// f = world.NewFixture("block5", colornames.Green, width, height)
	// f.Place(pixel.V(300, 100))
	// w.AddFixture(f)

	// width = 200
	// height = 10
	// f = world.NewFixture("block6", colornames.Green, width, height)
	// f.Place(pixel.V(300, 650))
	// w.AddFixture(f)
}
