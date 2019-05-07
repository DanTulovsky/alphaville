package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math"
	"time"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"golang.org/x/image/colornames"
)

// TargetSeekerBehavior moves in shortest path to the target
type TargetSeekerBehavior struct {
	DefaultBehavior
	target   Target
	qt       *Tree
	path     NodeList
	fullpath []pixel.Vec
	cost     int
	source   pixel.Vec
	// finder          graph.PathFinder // path finder function
	finder          PathFinder // path finder function
	turnsAtLocation int        // number of turns at current location
	targetsCaught   int64

	// TODO: Change this to be based on expected steps rather than wall time
	targetAcquireTime    time.Time     // when this target was acquired
	maxTargetAcquireTime time.Duration // max allowed time to get to the target
}

// NewTargetSeekerBehavior return a TargetSeekerBehavior
func NewTargetSeekerBehavior(f PathFinder) *TargetSeekerBehavior {
	return &TargetSeekerBehavior{
		DefaultBehavior: DefaultBehavior{
			name:        "target_seeker",
			description: "Travels in shortest path to target, if given, otherwise stands still.",
		},
		finder:               f,
		maxTargetAcquireTime: time.Second * time.Duration(utils.RandomInt(10, 20)),
	}
}

// RemainingTargetAcquireTime returns the remaining time to catch a target
func (b *TargetSeekerBehavior) RemainingTargetAcquireTime() time.Duration {
	return (b.maxTargetAcquireTime - time.Since(b.targetAcquireTime)).Round(time.Millisecond)
}

// MaxTargetAcquireTime returns the max time allowed to catch a target
func (b *TargetSeekerBehavior) MaxTargetAcquireTime() time.Duration {
	return b.maxTargetAcquireTime.Round(time.Millisecond)
}

// String returns ...
func (b *TargetSeekerBehavior) String() string {
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("physObject").Parse(
		`
Behavior
  Name: {{.Name}}	
    Desc: {{.Description}}	
    Target ({{.Target.ID}}): {{.Target.Location}}
    Turns At Location: {{.TurnsBlocked}}
    Remaining Time to Reach Target: {{.RemainingTargetAcquireTime}} of {{.MaxTargetAcquireTime}}
    Targets Caught: {{.TargetsCaught}}
    Path to Target: {{.FullPath}}
`)

	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}
	err = tmpl.Execute(buf, b)
	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}

	return buf.String()
}

// FullPath returns the full path to the current target
func (b *TargetSeekerBehavior) FullPath() []pixel.Vec {
	if b.target != nil {
		return b.fullpath
	}
	return []pixel.Vec{}
}

// populateMoveGraph creates a quadtree of the current world used to find the path from o (the seeker) to target
// https://cs.stanford.edu/people/eroberts/courses/soco/projects/1998-99/robotics/basicmotion.html
// https://www.dis.uniroma1.it/~oriolo/amr/slides/MotionPlanning1_Slides.pdf
func (b *TargetSeekerBehavior) populateMoveGraph(w *World) *Tree {
	// log.Printf("Populating move graph for %v", o.Name())

	// augmented fixtures, these are what we check collisions against
	// they are grown by 1/2 size of object on each side to account for movement
	cobjects, _ := w.CollisionObjectsExclude(b.parent) // must be all for now

	phys := b.parent.Phys()

	// add start and target to the quadtree
	s := phys.Location().Center()
	t := b.target.Location()
	start := pixel.R(s.X, s.Y, s.X, s.Y)
	target := pixel.R(t.X, t.Y, t.X, t.Y)

	startObj := NewBaseObject("start", colornames.Yellow, 0, 1000)
	startPhys := NewBaseObjectPhys(start, &startObj)
	startObj.SetPhys(startPhys)

	targetObj := NewBaseObject("target", colornames.Yellow, 0, 1000)
	targetPhys := NewBaseObjectPhys(target, &targetObj)
	targetObj.SetPhys(targetPhys)

	// Use own quadtree
	/////////////////////////////////////
	cobjects = append(cobjects, &startObj, &targetObj)

	// minimum size of rectangle side at which we stop splitting
	// based on the size of the target seeker
	minSize := math.Min(phys.Location().W(), phys.Location().H())

	// quadtree
	qtBounds := pixel.R(
		w.Ground.Phys().Location().Min.X+phys.Location().W()/2, w.Ground.Phys().Location().Max.Y+phys.Location().H()/2,
		w.X-phys.Location().W()/2, w.Y-phys.Location().H()/2)
	qt, err := NewTree(qtBounds, cobjects, minSize, phys.Location().Size())
	if err != nil {
		log.Fatalf("error creating quadtree: %v", err)
	}

	startNode, err := qt.Locate(startObj.Phys().Location().Center())
	if err != nil {
		log.Fatalf("%v", err)
	}
	targetNode, err := qt.Locate(targetObj.Phys().Location().Center())
	if err != nil {
		log.Fatalf("%v", err)
	}

	// flip the source and target nodes to be White so the path between them can be found
	startNode.SetColor(colornames.White)
	targetNode.SetColor(colornames.White)
	/////////////////////////////////////////////

	return qt

	// Use world quadtree
	// b.qt = w.QuadTree()
}

// SetTarget sets the target
func (b *TargetSeekerBehavior) SetTarget(t Target) {
	b.target = t
}

// Target returns the current target
func (b *TargetSeekerBehavior) Target() Target {
	return b.target
}

// TargetsCaught returns the current target
func (b *TargetSeekerBehavior) TargetsCaught() int64 {
	return b.targetsCaught
}

// TurnsBlocked returns the number of turns this object hasn't moved
func (b *TargetSeekerBehavior) TurnsBlocked() int {
	return b.turnsAtLocation
}

// isAtTarget returns true if any part of the object covers the target
func (b *TargetSeekerBehavior) isAtTarget(o Object) bool {

	// if utils.VecLen(o.Phys().Location().Center(), b.target.Bounds().Center()) < o.Speed() {
	if o.Phys().Location().Contains(b.target.Location()) {

		o.Notify(NewObjectEvent(
			fmt.Sprintf("[%v] found target [%v]", o.Name(), b.target.Name()), time.Now(),
			observer.EventData{Key: "target_found", Value: b.target.Name()}))
		b.target.Destroy()
		b.target = nil

		return true
	}
	return false
}

// Direction returns the next direction to travel and the target itself
func (b *TargetSeekerBehavior) Direction(w *World, o Object) (pixel.Vec, pixel.Vec) {
	// remove the current location from path
	// circle := pixel.C(o.Phys().Location().Center(), o.Speed()*2)

	// for len(b.path) > 0 && circle.Contains(b.path[0].Value().V) {
	for len(b.path) > 0 && utils.VecLen(o.Phys().Location().Center(), b.path[0].Bounds().Center()) < o.Speed() {
		// if len(b.path) > 0 && o.Phys().Location().Contains(b.path[0].Value().V) {

		b.source = b.path[0].Bounds().Center()
		b.path = append(b.path[:0], b.path[1:]...)
	}

	if len(b.path) == 0 {
		// log.Printf("path ran out...")
		return pixel.ZV, pixel.V(0, 0)
	}
	source := b.source
	// target is the next node in the path
	target := b.path[0].Bounds().Center()
	// current location of target seeker
	c := o.Phys().Location().Center()

	// log.Printf("From: %v; To: %v\n", source, target)
	// log.Printf("  Current location: %v (%v)", o.Phys().Location(), o.Phys().Location().Center())
	// log.Printf("  Target location: %v", target)
	// log.Printf("  at target?: %v", b.isAtTarget(o))
	var moves []pixel.Vec

	orient := graph.Orientation(source, target, c)
	// log.Printf("  orientation: %v", orient)

	if target.X > source.X {
		if utils.LineSlope(source, target) > 0 {
			switch orient {
			case 2:
				// if above, move x right
				moves = append(moves, pixel.V(1, 0))
			case 1:
				// if below, move y up
				moves = append(moves, pixel.V(0, 1))
			case 0:
				// if on the line, move in either direction
				moves = append(moves, pixel.V(0, 1))

			}
		}

		if utils.LineSlope(source, target) < 0 {
			switch orient {
			case 2:
				// if above, move y down
				moves = append(moves, pixel.V(0, -1))
			case 1:
				// if below, move x right
				moves = append(moves, pixel.V(1, 0))
			case 0:
				// if on the line, move in either direction
				moves = append(moves, pixel.V(0, -1))
			}

		}
	}

	if target.X < source.X {
		if utils.LineSlope(source, target) > 0 {
			switch orient {
			case 2:
				// if below, move x left
				moves = append(moves, pixel.V(-1, 0))
			case 1:
				// if above, move y down
				moves = append(moves, pixel.V(0, -1))
			case 0:
				// if on the line, move in either direction
				moves = append(moves, pixel.V(0, -1))
			}

		}

		if utils.LineSlope(source, target) < 0 {
			switch orient {
			case 2:
				// if below, move y up
				moves = append(moves, pixel.V(0, 1))
			case 1:
				// if above, move x left
				moves = append(moves, pixel.V(-1, 0))
			case 0:
				// if on the line, move in either direction
				moves = append(moves, pixel.V(0, 1))
			}
		}
	}

	if target.X == source.X {
		switch {
		// move y towards target
		case target.Y > source.Y:
			// move up
			moves = append(moves, pixel.V(0, 1))
		case target.Y < source.Y:
			// move down
			moves = append(moves, pixel.V(0, -1))
		}
	}

	if target.Y == source.Y {
		switch {
		// move x towards target
		case target.X > source.X:
			// move right
			moves = append(moves, pixel.V(1, 0))
		case target.X < source.X:
			// move left
			moves = append(moves, pixel.V(-1, 0))
		}
	}

	if len(moves) > 0 {
		return moves[utils.RandomInt(0, len(moves))], target
	}

	o.SetManualVelocity(pixel.ZV)
	return pixel.ZV, pixel.ZV

}

// FindPath returns the path and cost between start and target
func (b *TargetSeekerBehavior) FindPath(start, target pixel.Vec) (NodeList, int, error) {

	// log.Printf("looking for path from %v to %v", start, target)
	path, cost, err := b.finder.Path(b.qt, start, target)
	if err != nil {
		// log.Println(err)
		return nil, 0, err
	}

	return path, cost, err
}

// FindAndSetNewTarget grabs a new target from the world
func (b *TargetSeekerBehavior) FindAndSetNewTarget(w *World, o Object) error {

	var t Target
	var err error

	if t, err = w.GetTarget(); err != nil {
		return fmt.Errorf("error picking target: %v", err)
	}

	t.Register(b)
	b.SetTarget(t)
	// log.Printf("[%v] target [%v] acquired", b.parent.Name(), t.ID())

	b.recalculateMoveInfo(w, o)
	b.targetAcquireTime = time.Now()

	return nil
}

func (b *TargetSeekerBehavior) processTargetEvent(e *TargetEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "destroyed":
			if b.target == nil {
				return
			}
			if b.target.ID().String() == data.Value {
				// stop chasing destroyed targets
				// log.Printf("[%v] target [%v] destroyed need to pick another one", b.parent.Name(), data.Value)
				// b.target.Deregister(b)
				b.target = nil
			}
		}
	}
}

// recalculateMoveInfo recalculates the path for an existing target
func (b *TargetSeekerBehavior) recalculateMoveInfo(w *World, o Object) {
	phys := o.NextPhys()

	b.qt = b.populateMoveGraph(w)
	var err error
	startNode, err := b.qt.Locate(phys.Location().Center())
	if err != nil {
		log.Fatal(err)
	}
	targetNode, err := b.qt.Locate(b.target.Bounds().Center())
	if err != nil {
		log.Fatal(err)
	}

	b.fullpath = []pixel.Vec{}
	b.path, b.cost, err = b.FindPath(startNode.Bounds().Center(), targetNode.Bounds().Center())
	if err != nil {
		return
		// log.Printf("error finding path: %v", err)
	}

	for _, n := range b.path {
		b.fullpath = append(b.fullpath, n.Bounds().Center())
	}

	if len(b.fullpath) != 0 {
		// add current location
		b.fullpath = append([]pixel.Vec{phys.Location().Center()}, b.fullpath...)

		// add target
		b.fullpath = append(b.fullpath, b.target.Bounds().Center())
	}
	b.source = phys.Location().Center()
}

// Update implements the Behavior Update method
// In this method, execute the planned move in the NextPhys object
// If unable to do so due to any reason, change the Velocity, but do not move
// this turn, otherwise collision detection fails.
func (b *TargetSeekerBehavior) Update(w *World, o Object) {
	phys := o.NextPhys()

	if b.target == nil {
		if err := b.FindAndSetNewTarget(w, o); err != nil {
			// no target found
			return
		}
		return
	}

	if b.isAtTarget(o) {
		b.targetsCaught++
		return
	}

	// If too much wall clock time has passed, give up on this target and find another one
	if time.Since(b.targetAcquireTime) > b.maxTargetAcquireTime {
		log.Printf("[%v] Time spent (%v) to catch [%v] expired (max %v), trying another target...", b.parent.Name(), time.Since(b.targetAcquireTime), b.target.Name(), b.maxTargetAcquireTime)
		if err := b.FindAndSetNewTarget(w, o); err != nil {
			log.Printf("... but failed to find new target: %v", err)
		}
		return
	}

	// if stuck, redo the graph, but not too often to let things move out of the way
	if len(b.path) == 0 || b.turnsAtLocation > 8 {
		if b.turnsAtLocation%35 == 0 {
			b.recalculateMoveInfo(w, o)
		}
	}

	if len(phys.HaveCollisionsAt(w)) == 0 && !(phys.Vel() == pixel.ZV) {
		// move, checking collisions with world borders
		b.Move(w, o, phys.CollisionBordersVector(w, phys.Vel()))
		b.turnsAtLocation = 0
	} else {
		b.turnsAtLocation++
	}

	// unable to move via path for a long time, try random walk
	if b.turnsAtLocation > 50 {
		randx := float64(utils.RandomInt(-1, 2))
		randy := float64(utils.RandomInt(-1, 2))
		// if randx != 0 {
		// 	randy = 0
		// }
		v := pixel.V(randx, randy)
		phys.SetManualVelocityXY(v)
		return // delay actual move until next tick to avoid collision problems
	}

	// Setup next move
	d, target := b.Direction(w, o)
	phys.SetManualVelocityXY(d)

	// if moving takes us further away from the target than we currently are
	// just move directly on top of the target, if possible
	currentDistance := utils.VecLen(phys.Location().Center(), target)
	newDistance := utils.VecLen(phys.Location().Moved(phys.Vel()).Center(), target)

	if newDistance > currentDistance {
		v := target.Sub(phys.Location().Center())
		phys.SetVel(v)
	}
}

// Move moves the object
func (b *TargetSeekerBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}

// Draw draws any artifacts of the behavior
func (b *TargetSeekerBehavior) Draw(win *pixelgl.Window) {
	if b.target == nil {
		return
	}

	// draw the quadtree
	// drawTree, colorTree, drawText, drawObjects := true, true, false, true
	// b.qt.Draw(win, drawTree, colorTree, drawText, drawObjects)

	pathColor := b.parent.Color()

	if len(b.path) > 0 {
		// draw the path from current location
		l := b.parent.Phys().Location().Center()
		drawPath := make([]pixel.Vec, len(b.path)+2)
		drawPath[0] = l
		for i := 0; i < len(b.path); i++ {
			drawPath[i+1] = b.path[i].Bounds().Center()
		}
		drawPath[len(drawPath)-1] = b.target.Bounds().Center()
		DrawPath(win, drawPath, pathColor)
	}
	// draw the full path
	// DrawPath(win, b.fullpath, pathColor)
}

// Implement the EventObserver interface

// OnNotify runs when a notification is received
func (b *TargetSeekerBehavior) OnNotify(e observer.Event) {
	switch event := e.(type) {
	case nil:
		log.Printf("nil notification")
	case *TargetEvent:
		b.processTargetEvent(event)
	}
}

// Name returns the name of the object with this behavior
func (b *TargetSeekerBehavior) Name() string {
	return b.parent.Name()
}
