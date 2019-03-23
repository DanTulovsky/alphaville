package populate

import (
	"fmt"
	"log"
	"time"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

// RandomEllipses populates the world with N random objects
func RandomEllipses(w *world.World, n int) {

	var minRadius, maxRadius, minMass, maxMass, minSpeed, maxSpeed float64

	minRadius = 10
	maxRadius = 60
	minMass, maxMass = 1, 10
	minSpeed, maxSpeed = 1, 10

	for i := 0; i < n; i++ {
		randomColor := colornames.Map[colornames.Names[utils.RandomInt(0, len(colornames.Names))]]

		a := utils.RandomFloat64(minRadius, maxRadius+1)
		b := utils.RandomFloat64(minRadius, maxRadius+1)

		o := world.NewEllipseObject(
			fmt.Sprintf("%v", i),
			randomColor,
			utils.RandomFloat64(minSpeed, maxSpeed)/10, // speed
			utils.RandomFloat64(minMass, maxMass)/10,   // mass
			a,   // x radius
			b,   // y radius
			nil, // default behavior
		)

		w.AddObject(o)
	}
}

// RandomCircles populates the world with N random objects
func RandomCircles(w *world.World, n int) {

	var minRadius, maxRadius, minMass, maxMass, minSpeed, maxSpeed float64

	minRadius = 10
	maxRadius = 60
	minMass, maxMass = 1, 10
	minSpeed, maxSpeed = 1, 10

	for i := 0; i < n; i++ {
		randomColor := colornames.Map[colornames.Names[utils.RandomInt(0, len(colornames.Names))]]

		radius := utils.RandomFloat64(minRadius, maxRadius+1)

		o := world.NewCircleObject(
			fmt.Sprintf("%v", i),
			randomColor,
			utils.RandomFloat64(minSpeed, maxSpeed)/10, // speed
			utils.RandomFloat64(minMass, maxMass)/10,   // mass
			radius, // radius
			nil,    // default behavior
		)

		w.AddObject(o)
	}
}

// RandomRectangles populates the world with N random rectangular objects
func RandomRectangles(w *world.World, n int) {

	var minWidth, maxWidth, minHeight, maxHeight, minMass, maxMass, minSpeed, maxSpeed float64

	minWidth, maxWidth = 10, 81
	minHeight, maxHeight = 10, 81
	minMass, maxMass = 1, 10
	minSpeed, maxSpeed = 1, 10

	for i := 0; i < n; i++ {
		randomColor := colornames.Map[colornames.Names[utils.RandomInt(0, len(colornames.Names))]]

		width := utils.RandomFloat64(minWidth, maxWidth+1)
		height := utils.RandomFloat64(minHeight, maxHeight)

		o := world.NewRectObject(
			fmt.Sprintf("%v", i),
			randomColor,
			utils.RandomFloat64(minSpeed, maxSpeed)/10, // speed
			utils.RandomFloat64(minMass, maxMass)/10,   // mass
			width,  // width
			height, // height
			nil,    // default behavior
		)

		w.AddObject(o)
	}
}

// AddManualObject adds a manually controlled object to the world
func AddManualObject(w *world.World, width, height float64) {

	behavior := world.NewManualBehavior()

	o := world.NewRectObject(
		"manual",
		colornames.Red,
		4,      // speed
		1,      // mass
		width,  // width
		height, // height
		behavior,
	)

	w.AddObject(o)
	w.ManualControl = o
}

// AddGates adds gates to the world
func AddGates(w *world.World, coolDown time.Duration) {

	// add spawn gate
	gates := []world.Gate{
		{
			Location:      pixel.V(600, 600),
			Status:        world.GateOpen,
			SpawnCoolDown: 10 * time.Second,
			Radius:        20,
		},
		{
			Location:      pixel.V(200, 600),
			Status:        world.GateOpen,
			SpawnCoolDown: 2 * time.Second,
			Radius:        25,
		},
	}

	for _, g := range gates {
		gate := world.NewGate(g.Location, g.Status, g.SpawnCoolDown, g.Radius)

		// Register the world.stats object to receive notifications from the gate
		gate.EventNotifier.Register(w.Stats)
		if err := w.AddGate(gate); err != nil {
			log.Fatalf("error adding gate: %v", err)
		}
	}
}

// AddFixtures add fixtures to the world
func AddFixtures(w *world.World) {

	var width float64 = 10
	var height float64 = 100

	f := world.NewFixture("block1", colornames.Green, width, height)
	f.Place(pixel.V(300, w.Ground.Phys().Location().Max.Y+100))

	w.AddFixture(f)
}
