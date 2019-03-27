package world

import (
	"fmt"
	"log"
	"time"

	"gogs.wetsnow.com/dant/alphaville/observer"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
)

// World defines the world
type World struct {
	X, Y          float64  // size of the world
	Gates         []*Gate  // entrances into the world
	Objects       []Object // objects in the world
	targets       []Target // targets in the world that TargetSeekers hunt
	ManualControl Object   // this object is human controlled
	Ground        Object   // special, for now
	fixtures      []Object // walls, floors, rocks, etc...
	gravity       float64
	Stats         *Stats // world stats, an observer of events happening in the world

	observers []observer.EventObserver
}

// NewWorld returns a new worldof size x, y
func NewWorld(x, y float64, ground Object, gravity float64) *World {

	w := &World{
		Objects: []Object{},
		Gates:   []*Gate{},
		targets: []Target{},
		X:       x,
		Y:       y,
		Ground:  ground,
		gravity: gravity,
		Stats:   NewStats(),
		// EventNotifier: observer.NewEventNotifier(),
		ManualControl: NewNullObject(),
	}

	w.Register(w.Stats)
	w.Notify(w.NewWorldEvent(fmt.Sprintf("The world is created..."), time.Now()))
	return w
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

// Draw draws the world by callign each object's Draw()
func (w *World) Draw(win *pixelgl.Window) {
	w.Ground.Draw(win)

	for _, g := range w.Gates {
		g.Draw(win)
	}

	for _, f := range w.Fixtures() {
		f.Draw(win)
	}

	for _, t := range w.Targets() {
		t.Draw(win)
	}

	for _, o := range w.Objects {
		o.Draw(win)
		o.Behavior().Draw(win)
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

// Targets returns all the targets in the world
func (w *World) Targets() []Target {
	var targets []Target

	for _, t := range w.targets {
		targets = append(targets, t)
	}
	return targets
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

// AddTarget adds a new target to the world
func (w *World) AddTarget(t Target) error {
	if t.Location().X > w.X || t.Location().Y > w.Y || t.Location().X < 0 || t.Location().Y < 0 {
		return fmt.Errorf("Location %#v is outside the world bounds (%#v)", t.Location(), pixel.V(w.X, w.Y))
	}

	for _, target := range w.Targets() {
		if t.Location() == target.Location() {
			return fmt.Errorf("target at %v already exists (%v)", t.Location(), t.Name())
		}
	}
	w.targets = append(w.targets, t)

	// register those who should be notified of events
	t.Register(w.Stats)
	t.Register(w)

	t.Notify(NewTargetEvent(fmt.Sprintf("target [%v] created", t), time.Now(),
		observer.EventData{Key: "created", Value: t.ID().String()}))
	return nil
}

// RemoveTarget removes a target from the world
func (w *World) RemoveTarget(uuid string) {
	targets := []Target{}
	for _, t := range w.Targets() {
		if t.ID().String() != uuid {
			targets = append(targets, t)
		}
	}
	w.targets = targets
}

// AvailableTargets returns a list of available targets in the world
func (w *World) AvailableTargets() []Target {
	return w.Targets()
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

	g.Register(w.Stats)
	g.Register(w)

	g.Notify(NewGateEvent(fmt.Sprintf("gate [%v] created", g), time.Now(),
		observer.EventData{Key: "created", Value: g.Name()}))
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
	g.Notify(NewGateEvent(
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
		if g.Reserve(o) == nil {
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
	w.Notify(w.NewWorldEvent(fmt.Sprint("The world dies..."), time.Now()))
	w = nil
}

// ShowStats dumps the world stats to stdout
func (w *World) ShowStats() {
	fmt.Printf("%v", w.Stats)
}

// ObjectClicked returns the object at coordinates v
func (w *World) ObjectClicked(v pixel.Vec) (Object, error) {
	for _, o := range append(w.SpawnedObjects()) {
		if o.Phys().Location().Contains(v) {
			return o, nil
		}
	}

	for _, o := range append(w.Fixtures()) {
		if o.Phys().Location().Contains(v) {
			return o, nil
		}
	}

	for _, g := range append(w.Gates) {
		c := pixel.C(g.Location, g.Radius)
		if c.Contains(v) {
			return g, nil
		}
	}
	return nil, fmt.Errorf("no object at %v", v)
}
