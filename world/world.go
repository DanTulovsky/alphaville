package world

import (
	"fmt"
	"log"
	"time"

	"gogs.wetsnow.com/dant/alphaville/observer"

	"github.com/faiface/pixel"
)

// World defines the world
type World struct {
	X, Y          float64  // size of the world
	Gates         []*Gate  // entrances into the world
	Objects       []Object // objects in the world
	ManualControl Object   // this object is human controlled
	Ground        Object   // special, for now
	fixtures      []Object // walls, floors, rocks, etc...
	gravity       float64
	Stats         *Stats // world stats, an observer of events happening in the world
	EventNotifier observer.EventNotifier
}

// NewWorld returns a new worldof size x, y
func NewWorld(x, y float64, ground Object, gravity float64) *World {

	w := &World{
		Objects:       []Object{},
		Gates:         []*Gate{},
		X:             x,
		Y:             y,
		Ground:        ground,
		gravity:       gravity,
		Stats:         NewStats(),
		EventNotifier: observer.NewEventNotifier(),
		ManualControl: NewNullObject(),
	}

	w.EventNotifier.Register(w.Stats)
	w.EventNotifier.Notify(w.NewWorldEvent(fmt.Sprintf("The world is created..."), time.Now()))
	return w
}

type worldEvent struct {
	observer.BaseEvent
	worldType Type
}

// NewWorldEvent create a new world event
func (w *World) NewWorldEvent(d string, t time.Time, data ...observer.EventData) observer.Event {
	e := &worldEvent{
		worldType: worldType,
	}
	e.SetData(data)
	e.SetDescription(d)
	e.SetTime(t)

	return e
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
	// update movable objects
	for _, o := range w.SpawnedObjects() {
		o.Update(w)
	}

	// update fixtures
	for _, o := range w.Fixtures() {
		o.Update(w)
	}
}

// NextTick moves the world to the next state
func (w *World) NextTick() {
	// After update, swap the state of all objects at once
	for _, o := range w.SpawnedObjects() {
		o.SwapNextState()
	}
}

// Fixtures returns all the fixtures in the world
func (w *World) Fixtures() []Object {
	var fs []Object

	for _, f := range w.fixtures {
		fs = append(fs, f)
	}
	return fs
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

// CollisionObjects returns all objects for which to check collisions.
func (w *World) CollisionObjects() []Object {
	return append(w.SpawnedObjects(), w.Fixtures()...)
}

// AddObject adds a new object to the world
func (w *World) AddObject(o Object) {
	w.Objects = append(w.Objects, o)
}

// AddFixture adds a new fixture to the world
func (w *World) AddFixture(o Object) {
	w.fixtures = append(w.fixtures, o)
}

// AddGate adds a new gate to the world
func (w *World) AddGate(g *Gate) error {
	if g.Location.X > w.X || g.Location.Y > w.Y || g.Location.X < 0 || g.Location.Y < 0 {
		return fmt.Errorf("Location %#v is outside the world bounds (%#v)", g.Location, pixel.V(w.X, w.Y))
	}

	for _, gate := range w.Gates {
		if g.Location == gate.Location {
			return fmt.Errorf("gate at %v already exists", g.Location)
		}
	}
	w.Gates = append(w.Gates, g)
	g.EventNotifier.Notify(w.NewWorldEvent(fmt.Sprintf("gate [%v] created", g), time.Now()))
	return nil
}

// SpawnObject tries to grab a gate and spawn, if area around is available
// We also create the Phys() of the object here
func (w *World) SpawnObject(o Object) error {

	g, err := w.ReserveGate(o)
	if err != nil {
		return err
	}

	phys := NewBaseObjectPhys(o.BoundingBox(g.Location), o)

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

	// phys.SetVel(pixel.V(o.Speed(), w.gravity*o.Mass()))

	// Don't set velocity for manual objects
	switch o.Behavior().(type) {
	case *ManualBehavior:
	default:
		phys.SetVel(pixel.V(o.Speed(), 0))
	}
	phys.SetCurrentMass(o.Mass())

	o.SetPhys(phys)
	o.SetNextPhys(o.Phys().Copy())

	g.Release()
	g.EventNotifier.Notify(w.NewWorldEvent(
		fmt.Sprintf(
			"object [%v] spawned", o.Name()), time.Now(),
		observer.EventData{Key: "spawn", Value: fmt.Sprintf("%T", o)}))

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

// End destroys the world
func (w *World) End() {
	w.EventNotifier.Notify(w.NewWorldEvent(fmt.Sprint("The world dies..."), time.Now()))
	w = nil
}

// ShowStats dumps the world stats to stdout
func (w *World) ShowStats() {
	fmt.Printf("%v", w.Stats)
}
