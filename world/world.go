package world

import (
	"fmt"
	"log"
	"time"

	"gogs.wetsnow.com/dant/alphaville/behavior"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

// World defines the world
type World struct {
	X, Y    float64  // size of the world
	Gates   []*Gate  // entrances into the world
	Objects []Object // objects in the world
	Ground  Object
	gravity float64
	stats   *Stats // world stats, an observer of events happening in the world
	Atlas   *text.Atlas
}

// NewWorld returns a new worldof size x, y
func NewWorld(x, y float64, ground Object, gravity float64) *World {
	return &World{
		Objects: []Object{},
		Gates:   []*Gate{},
		X:       x,
		Y:       y,
		Ground:  ground,
		gravity: gravity,
		stats:   NewStats(),
		Atlas:   text.NewAtlas(basicfont.Face7x13, text.ASCII),
	}
}

// SpawnAllNew spanws all new objects
func (w *World) SpawnAllNew() {
	for _, o := range w.UnSpawnedObjects() {

		if !o.IsSpawned() {
			if err := w.SpawnObject(o); err != nil {
				continue
			}
		}
	}
}

// Update updates all the objects in the world to their next state
func (w *World) Update() {
	for _, o := range w.SpawnedObjects() {
		o.Update(w)
	}
}

// NextTick moves the world to the next state
func (w *World) NextTick() {

	// After update, swap the state of all objects at once
	for _, o := range w.SpawnedObjects() {
		o.SwapNextState()

		// Remove gate reservations
		// for _, g := range w.Gates {
		// 	if g.Reserved {
		// 		g.Release()
		// 	}
		// }
	}
}

// SpawnedObjects returns all the spawned objects in the world
func (w *World) SpawnedObjects() []Object {
	var objs []Object
	for _, o := range w.Objects {
		if o.IsSpawned() {
			objs = append(objs, o)
		}
	}

	return objs
}

// UnSpawnedObjects returns all the unspawned objects in the world
func (w *World) UnSpawnedObjects() []Object {
	var objs []Object
	for _, o := range w.Objects {
		if !o.IsSpawned() {
			objs = append(objs, o)
		}
	}

	return objs
}

// AddObject adds a new object to the world
func (w *World) AddObject(o Object) {
	w.Objects = append(w.Objects, o)
}

// AddGate adds a new gate to the world
func (w *World) AddGate(g *Gate) error {
	for _, gate := range w.Gates {
		if g.Location == gate.Location {
			return fmt.Errorf("gate at %v already exists", g.Location)
		}
	}
	w.Gates = append(w.Gates, g)
	return nil
}

// NewGate creates a new gate in the world
func (w *World) NewGate(l pixel.Vec, s gateStatus, coolDown time.Duration, radius float64, atlas *text.Atlas) error {
	if l.X > w.X || l.Y > w.Y || l.X < 0 || l.Y < 0 {
		return fmt.Errorf("Location %#v is outside the world bounds (%#v)", l, pixel.V(w.X, w.Y))
	}

	g := &Gate{
		Location:      l,
		Status:        s,
		Reserved:      false,
		SpawnCoolDown: coolDown,
		Atlas:         atlas,
		Radius:        radius,
		eventNotifier: behavior.NewEventNotifier(),
	}

	// Register the world.stats object to receive notifications from the gate
	g.eventNotifier.Register(w.stats)

	return w.AddGate(g)
}

// SpawnObject tries to grab a gate and spawn, if area around is available
// We also create the Phys() of the object here
func (w *World) SpawnObject(o Object) error {

	g, err := w.ReserveGate(o)
	if err != nil {
		return err
	}

	phys := NewBaseObjectPhys(o.BoundingBox(g.Location))

	// If below ground, move up
	if phys.Location().Min.Y < w.Ground.Phys().Location().Max.Y {
		phys.SetLocation(phys.Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-phys.Location().Min.Y)))
	}

	// Spawn happens after everything already moved, so simply check for intersections here
	for _, other := range w.SpawnedObjects() {
		if phys.Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			return fmt.Errorf("not spawning, would intersect with %v", other.Name())
		}
	}

	phys.SetVel(pixel.V(o.Speed(), 0))
	phys.SetCurrentMass(o.Mass())

	o.SetPhys(phys)
	o.SetNextPhys(o.Phys().Copy())

	g.Release()
	return nil
}

// ReserveGate returns (and reserves) a gate to be used by an object
// An error is returned if there are no available gates
func (w *World) ReserveGate(o Object) (*Gate, error) {
	// make sure only one gate is reserved by an object
	for _, g := range w.Gates {
		if g.ReservedBy == o.ID() {
			return g, nil // returned gate already reserved
		}
	}

	for _, g := range w.Gates {
		if g.Reserve(o.ID()) == nil {
			return g, nil
		}
	}

	return nil, fmt.Errorf("no avalable gates")
}

// CheckIntersect checks if any objects in the world intersect and prints an error.
func (w *World) CheckIntersect() {
	for _, o := range w.Objects {
		for _, other := range w.Objects {
			if o.ID() == other.ID() {
				continue // skip yourself
			}
			if o.Phys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
				log.Printf("%#v intersects with %#v", o, other)
			}

		}
	}
}
