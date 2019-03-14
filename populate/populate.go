package populate

import (
	"fmt"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"gogs.wetsnow.com/dant/alphaville/world"
	"golang.org/x/image/colornames"
)

// Static puts some objects in the world
func Static(w *world.World) {
	var objs []*world.Object

	objs = append(objs,
		world.NewObject(
			"one",
			colornames.Red,
			2,
			1,
			20,
			20,
			world.NewObjectPhys(),
			w.Atlas))

	objs = append(objs,
		world.NewObject(
			"two",
			colornames.Blue,
			1,
			3,
			20,
			20,
			world.NewObjectPhys(),
			w.Atlas))

	var x float64
	for _, o := range objs {
		if o.IY == 0 {
			// place randomly, avoid intersection
			o.IY = utils.RandomFloat64(w.Ground.Phys.Rect.Max.Y, w.Y-o.H)
		}
		if o.IX == 0 {
			// place randomly, avoid intersection
			o.IX = x
			x += o.W + 1
		}
		// set bounding rectangle based on size and location
		o.Phys.Rect = pixel.R(o.IX, o.IY, o.W+o.IX, o.H+o.IY)

		// set velocity vector
		o.Phys.Vel = pixel.V(o.Speed, 0)

		// set current mass based on initial mass
		o.Phys.CurrentMass = o.Mass

		w.Objects = append(w.Objects, o)
	}

}

// Random populates the world with N random objects
func Random(w *world.World, n int) {

	var objs []*world.Object
	var x float64

	for i := 0; i < n; i++ {
		randomColor := colornames.Map[colornames.Names[utils.RandomInt(0, len(colornames.Names))]]

		o := world.NewObject(
			fmt.Sprintf("%v", i),
			randomColor,
			utils.RandomFloat64(10, 20)/10, // speed
			utils.RandomFloat64(1, 10)/10,  // mass
			utils.RandomFloat64(10, 81),    // width
			utils.RandomFloat64(10, 81),    // height
			world.NewObjectPhys(),
			w.Atlas,
		)

		if o.IY == 0 {
			// place randomly, avoid intersection
			o.IY = utils.RandomFloat64(w.Ground.Phys.Rect.Max.Y, w.Y-o.H)
		}
		if o.IX == 0 {
			// place randomly, avoid intersection
			o.IX = x
			x += o.W + 1
		}
		// set bounding rectangle based on size and location
		o.Phys.Rect = pixel.R(o.IX, o.IY, o.W+o.IX, o.H+o.IY)

		// set velocity vector
		o.Phys.Vel = pixel.V(o.Speed, 0)

		// set current mass based on initial mass
		o.Phys.CurrentMass = o.Mass

		objs = append(objs, o)

	}

	w.Objects = objs
}
