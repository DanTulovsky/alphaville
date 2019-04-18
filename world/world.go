package world

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/jroimartin/gocui"
)

// World defines the world
type World struct {
	name    string
	X, Y    float64  // size of the world
	Gates   []*Gate  // entrances into the world
	Objects []Object // objects in the world

	// qt keeps track of all the collidable objects in the world
	qt *Tree

	targets        []Target // targets in the world that TargetSeekers hunt
	removeTargets  []Target // targets to be removed next turn
	ManualControl  Object   // this object is human controlled
	Ground         Object   // special, for now
	fixtures       []Object // walls, floors, rocks, etc...
	gravity        float64
	Stats          *Stats // world stats, an observer of events happening in the world
	MaxObjectSpeed float64

	MinObjectSide float64 // minimum side of any object in the world

	observers []observer.EventObserver

	debug   *DebugConfig
	console *gocui.Gui
}

// NewWorld returns a new world of size x, y
func NewWorld(x, y float64, ground Object, gravity float64, maxSpeed float64, debug *DebugConfig, console *gocui.Gui) *World {

	w := &World{
		Objects: []Object{},
		Gates:   []*Gate{},
		targets: []Target{},
		X:       x,
		Y:       y,
		Ground:  ground,
		gravity: gravity,
		// EventNotifier: observer.NewEventNotifier(),
		ManualControl:  NewNullObject(),
		MaxObjectSpeed: maxSpeed,
		MinObjectSide:  20,
		debug:          debug,
		console:        console,
	}
	qt, err := NewTree(pixel.R(0, 0, x, y), []Object{}, w.MinObjectSide, pixel.ZV)
	if err != nil {
		log.Fatalf("cannot create world: %v", err)
	}

	w.qt = qt
	w.Stats = NewStats(w.ConsoleO())

	w.Register(w.Stats)
	w.Notify(w.NewWorldEvent(fmt.Sprintf("The world is created..."), time.Now()))
	return w
}

// String ...
func (w *World) String() string {
	output := bytes.NewBufferString("")

	fmt.Fprintln(output, "")
	fmt.Fprintf(output, "World: %v\n", w.name)
	fmt.Fprintf(output, "  Size: [%v, %v]\n", w.X, w.Y)
	fmt.Fprintf(output, "  QT:\n  %v\n", w.qt)
	fmt.Fprintln(output, "")

	return output.String()
}

// ConsoleI returns the input console
func (w *World) ConsoleI() *gocui.View {
	v, _ := w.console.View("input")
	return v
}

// ConsoleO returns the output console
func (w *World) ConsoleO() *gocui.View {
	v, err := w.console.View("output")
	if err != nil {
		log.Fatalln(err)
	}
	return v
}

// QuadTree returns the world quadtree
func (w *World) QuadTree() *Tree {
	return w.qt
}

// SpawnAllNew spawns all new objects
func (w *World) SpawnAllNew() {
	for _, o := range w.UnSpawnedObjects() {

		if !o.IsSpawned() {
			if gate, err := w.SpawnObject(o); err != nil {
				if gate != nil {
					gate.Release()
				}
				continue
			}
		}
	}
}

// Draw draws the world by calling each object's Draw()
func (w *World) Draw(win *pixelgl.Window) {
	w.Ground.Draw(win)

	for _, g := range w.Gates {
		g.Draw(win)
	}

	for _, f := range w.Fixtures() {
		f.Draw(win)
	}

	for _, o := range w.Objects {
		o.Behavior().Draw(win)
		o.Draw(win)
	}

	for _, t := range w.Targets() {
		t.Draw(win)
	}

	w.qt.Draw(win, w.debug.QT.DrawTree, w.debug.QT.ColorTree, w.debug.QT.DrawText, w.debug.QT.DrawObjects)

}

// Update updates all the objects in the world to their next state
func (w *World) Update() {
	w.Cleanup()

	// TODO: Replace with updating an existing tree when possible
	var err error
	cobjects, _ := w.CollisionObjects()
	w.qt, err = NewTree(pixel.R(0, 0, w.X, w.Y), cobjects, w.MinObjectSide, pixel.ZV)
	if err != nil {
		log.Fatalf("error creating world qt: %v", err)
	}

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

	targets = append(targets, w.targets...)
	return targets
}

// TargetObjects returns all the targets in the world as objects
func (w *World) TargetObjects() []Object {
	targets := []Object{}

	for _, t := range w.targets {
		targets = append(targets, t)
	}
	return targets
}

// Fixtures returns all the fixtures in the world
func (w *World) Fixtures() []Object {
	var fs []Object

	fs = append(fs, w.fixtures...)
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

// UnSpawnedObjects returns all the ready to spawn objects in the world
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
func (w *World) CollisionObjects() ([]Object, error) {
	return append(w.SpawnedObjects(), w.Fixtures()...), nil
}

// CollisionObjectsExclude returns all collission objects, excluding the passed in one
// This is used in making QuadTrees for the TargetSeekers. Exclude the TS itself so it can move freely.
// We add a point node in its place (the rest of the objects are augmented by its size).
func (w *World) CollisionObjectsExclude(o Object) ([]Object, error) {
	objects := []Object{}
	for _, other := range append(w.SpawnedObjects(), w.Fixtures()...) {
		if o.ID() == other.ID() {
			continue
		}
		objects = append(objects, other)
	}
	return objects, nil
}

// CollisionObjectsWith returns all objects for which to check collisions for the given object
func (w *World) CollisionObjectsWith(o Object) ([]Object, error) {
	// Find the quadrant in w.qt that includes center of o
	node, err := w.qt.Locate(o.Phys().Location().Center())
	if err != nil {
		return nil, err
	}
	// walk up the parent objects until node fully encloses o, with no intersections
	isect := true

	// TODO: But also need to account for objects that might go into the node?
	for isect {
		if utils.RectContains(node.Bounds(), o.Phys().Location()) {
			isect = false
		} else {
			node = node.Parent()
		}
	}

	return node.Objects(), nil
}

// checkObjectValid checks if the object is valid to be added to the world
func (w *World) checkObjectValid(o Object) error {
	sizex, sizey := o.BoundingBox(pixel.ZV).Size().XY()
	if math.Min(sizex, sizey) < w.MinObjectSide {
		return fmt.Errorf("object too small; min side is %v, object is: [%v, %v]", w.MinObjectSide, sizex, sizey)
	}
	return nil
}

// AddObject adds a new object to the world
func (w *World) AddObject(o Object) error {
	if err := w.checkObjectValid(o); err != nil {
		return err
	}
	w.Objects = append(w.Objects, o)
	return nil
}

// AddFixture adds a new fixture to the world
func (w *World) AddFixture(o Object) error {
	if err := w.checkObjectValid(o); err != nil {
		return err
	}
	w.fixtures = append(w.fixtures, o)
	return nil
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
func (w *World) RemoveTarget(remove Target) {
	targets := []Target{}
	for _, t := range w.Targets() {
		if t.ID() != remove.ID() {
			targets = append(targets, t)
		}
	}
	w.targets = targets
}

// RemoveOldTargets removes any targets slated for deletion
func (w *World) RemoveOldTargets() {
	if len(w.removeTargets) == 0 {
		return
	}

	for _, t := range w.removeTargets {
		w.RemoveTarget(t)
	}

	w.removeTargets = []Target{}
}

// Cleanup runs every tick and does general cleanup
func (w *World) Cleanup() {
	w.RemoveOldTargets()
}

// RegisterTargetRemoval marks this target for removal next turn
func (w *World) RegisterTargetRemoval(uuid string) {
	for _, t := range w.Targets() {
		if t.ID().String() == uuid {
			w.removeTargets = append(w.removeTargets, t)
			t.SetAvailable(false)
		}
	}
}

// GetTarget returns an available target
func (w *World) GetTarget() (Target, error) {
	if len(w.targets) == 0 {
		return nil, fmt.Errorf("no available targets")
	}

	var targets []Target

	for _, t := range w.targets {
		if t.Available() {
			targets = append(targets, t)
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no available targets")
	}

	return targets[utils.RandomInt(0, len(targets))], nil
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
// We also create the Phys() of the object here, it returns the reserved gate
func (w *World) SpawnObject(o Object) (*Gate, error) {

	g, err := w.ReserveGate(o)
	if err != nil {
		return nil, err
	}

	phys := NewBaseObjectPhys(o.BoundingBox(g.Location), o)

	// If below ground, move up
	if phys.Location().Min.Y < w.Ground.Phys().Location().Max.Y {
		phys.SetLocation(phys.Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-phys.Location().Min.Y)))
	}

	// Spawn happens after everything already moved, so simply check for intersections here
	for _, other := range append(w.SpawnedObjects(), w.Fixtures()...) {
		if phys.Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			return g, fmt.Errorf("not spawning, would intersect with %v", other.Name())
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

	return g, nil
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

	r := rand.New(rand.NewSource(time.Now().Unix()))

	for _, i := range r.Perm(len(w.Gates)) {
		g := w.Gates[i]
		if g.Reserve(o) == nil {
			return g, nil
		}
	}

	return nil, fmt.Errorf("no available gates")
}

// CheckIntersect checks if any objects in the world intersect and prints an error.
func (w *World) CheckIntersect() {
	for _, o := range w.Objects {
		for _, other := range w.Objects {
			if o.ID() == other.ID() {
				continue // skip yourself
			}
			if o.Phys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
				fmt.Fprintf(w.ConsoleO(), "%#v intersects with %#v", o, other)
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
	fmt.Fprintf(w.ConsoleO(), "%v\n", w.Stats)
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

	for _, t := range append(w.Targets()) {
		if t.Circle().Contains(v) {
			return t, nil
		}
	}
	return nil, fmt.Errorf("no object at %v", v)
}
