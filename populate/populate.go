package populate

import (
	"fmt"
	"log"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

// Static puts some objects in the world
// func Static(w *world.World) {
// 	var objs []*world.RectObject

// 	objs = append(objs,
// 		world.NewRectObject(
// 			"one",
// 			colornames.Red,
// 			2,
// 			1,
// 			20,
// 			20,
// 			world.NewRectObjectPhys(pixel.R(0, 0, 0, 0)),
// 			w.Atlas))

// 	objs = append(objs,
// 		world.NewRectObject(
// 			"two",
// 			colornames.Blue,
// 			1,
// 			3,
// 			20,
// 			20,
// 			world.NewRectObjectPhys(pixel.R(0, 0, 0, 0)),
// 			w.Atlas))

// 	var x float64
// 	for _, o := range objs {
// 		if o.IY == 0 {
// 			// place randomly, avoid intersection
// 			o.IY = utils.RandomFloat64(w.Ground.Phys().Location().Max.Y, w.Y-o.H)
// 		}
// 		if o.IX == 0 {
// 			// place randomly, avoid intersection
// 			o.IX = x
// 			x += o.W + 1
// 		}
// 		// set bounding rectangle based on size and location
// 		o.Phys().SetLocation(pixel.R(o.IX, o.IY, o.W+o.IX, o.H+o.IY))

// 		// set velocity vector
// 		o.Phys().SetVel(pixel.V(o.Speed, 0))

// 		// set current mass based on initial mass
// 		o.Phys().SetCurrentMass(o.Mass)

// 		w.AddObject(o)
// 	}

// }

// RandomEllipses populates the world with N random objects
// ystart specifies the lowest point these objects appear
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
			a, // x radius
			b, // y radius
			w.Atlas,
		)

		w.AddObject(o)
	}
}

// RandomCircles populates the world with N random objects
// ystart specifies the lowest point these objects appear
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
			w.Atlas,
		)

		w.AddObject(o)
	}
}

// RandomRectangles populates the world with N random rectangular objects
// ystart specifies the lowest point these objects appear
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
			w.Atlas,
		)

		w.AddObject(o)
	}

}

// AddGates adds gates to the world
func AddGates(w *world.World, coolDown time.Duration, atlas *text.Atlas) {

	// add spawn gate
	gates := []world.Gate{
		{
			Location:      pixel.V(600, 600),
			Status:        world.GateOpen,
			SpawnCoolDown: coolDown,
			Radius:        20,
			Atlas:         atlas,
		},
		// {
		// 	Location:      pixel.V(600, 600),
		// 	Status:        world.GateOpen,
		// 	SpawnCoolDown: coolDown,
		// },
	}

	for _, g := range gates {
		if err := w.NewGate(g.Location, g.Status, g.SpawnCoolDown, g.Radius, g.Atlas); err != nil {
			log.Fatalf("failed to create gate: %v", err)
		}
	}
}
