package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math"
	"time"

	"gogs.wetsnow.com/dant/alphaville/quadtree"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/pixelgl"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Behavior is the interface for all behaviors
type Behavior interface {
	Description() string
	Draw(win *pixelgl.Window)
	Name() string
	Parent() Object
	SetParent(Object)
	Update(*World, Object)
}

// DefaultBehavior is the default implementation of Behavior
type DefaultBehavior struct {
	description string
	name        string
	parent      Object
}

// NewDefaultBehavior return a DefaultBehavior
func NewDefaultBehavior() *DefaultBehavior {
	return &DefaultBehavior{
		description: "",
		name:        "default_behavior",
	}
}

// String returns ...
func (b *DefaultBehavior) String() string {
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("physObject").Parse(
		`
Behavior
  Name: {{.Name}}	
  Desc: {{.Description}}	
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

// Name returns the name of the behavior
func (b *DefaultBehavior) Name() string {
	return b.name
}

// Parent returns the parent object of the behavior
func (b *DefaultBehavior) Parent() Object {
	return b.parent
}

// SetParent returns the parent object of the behavior
func (b *DefaultBehavior) SetParent(p Object) {
	b.parent = p
}

// Description returns the name of the behavior
func (b *DefaultBehavior) Description() string {
	return b.description
}

// Update executes the next world step for the object
// It updates the NextPhys() of the object for next step based on the encoded behavior
func (b *DefaultBehavior) Update(w *World, o Object) {

	// Movement and location are set in the NextPhys object
	phys := o.NextPhys()

	// check if object should rise or fall, these checks not based on collisions
	// if anything changes, leave actual movement until next turn, otherwise
	// collision detection gets confused
	if b.changeVerticalDirection(w, o) {
		return
	}

	// check collisions and adjust movement parameters
	// if a collision is detected, no movement happens this round
	if b.HandleCollisions(w, o) {
		return
	}

	// no collisions detected, move
	b.Move(w, o, pixel.V(phys.Vel().X, phys.Vel().Y))
}

// changeVerticalDirection updates the vertical direction if needed
func (b *DefaultBehavior) changeVerticalDirection(w *World, o Object) bool {
	phys := o.NextPhys()
	currentY := phys.Vel().Y

	if phys.IsAboveGround(w) {
		// fall speed based on mass and gravity
		new := phys.Vel()
		new.Y = w.gravity * phys.CurrentMass()
		phys.SetVel(new)

		if phys.Vel().X != 0 {
			v := phys.PreviousVel()
			v.X = phys.Vel().X
			phys.SetPreviousVel(v)

			v = phys.Vel()
			v.X = 0
			phys.SetVel(v)
		}
	}

	if phys.IsZeroMass() {
		// rise speed based on mass and gravity
		v := phys.Vel()
		v.Y = -1 * w.gravity * o.Mass()
		phys.SetVel(v)

		if phys.Vel().X != 0 {
			v = phys.PreviousVel()
			v.X = phys.Vel().X
			phys.SetPreviousVel(v)

			v = phys.Vel()
			v.X = 0
			phys.SetVel(v)
		}
	}
	// something was changed
	if currentY != phys.Vel().Y {
		return true
	}

	return false
}

// HandleCollisions returns true if o has any collisions
// it adjusts the physical properties of o to avoid the collision
func (b *DefaultBehavior) HandleCollisions(w *World, o Object) bool {
	phys := o.NextPhys()

	clocations := phys.HaveCollisionsAt(w)

	for _, clocation := range clocations {

		switch {
		case phys.MovingDown() && clocation == "below":
			b.avoidCollisionBelow(phys)
			return true
		case phys.MovingUp() && clocation == "above":
			b.avoidCollisionAbove(phys, w)
			return true
		case phys.MovingRight() && clocation == "right":
			b.avoidCollisionRight(phys)
			return true
		case phys.MovingLeft() && clocation == "left":
			b.avoidCollisionLeft(phys)
			return true
		}
	}
	return false
}

// avoidCollisionBelow changes o to avoid collision with an object below while moving down
func (b *DefaultBehavior) avoidCollisionBelow(phys ObjectPhys) {

	// avoid collision by stopping the fall and rising again
	phys.SetCurrentMass(0)
	v := phys.Vel()
	v.Y = 0
	phys.SetVel(v)
}

// avoidCollisionAbove changes o to avoid collision with an object above while moving up
func (b *DefaultBehavior) avoidCollisionAbove(phys ObjectPhys, w *World) {

	phys.SetCurrentMass(phys.ParentObject().Mass())
	v := phys.Vel()
	v.Y = 0
	// if on ground, Y is now 0 and X is 0 from before, reset X movement
	if phys.OnGround(w) {
		v.X = phys.PreviousVel().X
	}
	phys.SetVel(v)
}

// ChangeHorizontalDirection changes the horizontal direction of the object to the opposite of current
func (b *DefaultBehavior) ChangeHorizontalDirection(phys ObjectPhys) {
	v := phys.Vel()
	v.X = -1 * v.X
	phys.SetVel(v)
}

// avoidHorizontalCollision changes the object to avoid a horizontal collision
func (b *DefaultBehavior) avoidHorizontalCollision(phys ObjectPhys) {

	// Going to bump, 50/50 chance of rising up or changing direction
	if utils.RandomInt(0, 100) > 50 {
		phys.SetCurrentMass(0)
	} else {
		b.ChangeHorizontalDirection(phys)
	}
}

// avoidCollisionLeft changes o to avoid a collision on the left
func (b *DefaultBehavior) avoidCollisionLeft(phys ObjectPhys) {
	b.avoidHorizontalCollision(phys)
}

// avoidCollisionRight changes o to avoid a collision on the right
func (b *DefaultBehavior) avoidCollisionRight(phys ObjectPhys) {
	b.avoidHorizontalCollision(phys)
}

// Move moves the object by Vector, accounting for world boundaries
func (b *DefaultBehavior) Move(w *World, o Object, v pixel.Vec) {
	phys := o.NextPhys()

	// move vector that takes into account border collisions
	mv := phys.CollisionBordersVector(w, v)

	// TODO: Clean this so code is not duplicated with above function
	switch {
	case phys.MovingLeft() && phys.Location().Min.X+phys.Vel().X <= 0:
		// left border
		b.ChangeHorizontalDirection(phys)

	case phys.MovingRight() && phys.Location().Max.X+phys.Vel().X >= w.X:
		// right border
		b.ChangeHorizontalDirection(phys)

	case phys.MovingDown() && phys.Location().Min.Y+phys.Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		v := phys.Vel()
		v.Y = 0
		v.X = phys.PreviousVel().X
		phys.SetVel(v)

	case phys.MovingUp() && phys.Location().Max.Y+phys.Vel().Y >= w.Y && phys.Vel().Y > 0:
		// stop at ceiling if going up
		v := phys.Vel()
		v.Y = 0
		phys.SetVel(v)
		phys.SetCurrentMass(o.Mass())

	}
	phys.SetLocation(phys.Location().Moved(mv))
}

// Draw draws any artifacts of the behavior
func (b *DefaultBehavior) Draw(win *pixelgl.Window) {

}

// ManualBehavior is human controlled
type ManualBehavior struct {
	DefaultBehavior
}

// NewManualBehavior return a ManualBehavior
func NewManualBehavior() *ManualBehavior {
	b := &ManualBehavior{}
	b.name = "manual_behavior"
	b.description = "Controlled by a human."
	return b
}

// Update implements the Behavior Update method
func (b *ManualBehavior) Update(w *World, o Object) {
	phys := o.NextPhys()

	if len(phys.HaveCollisionsAt(w)) == 0 {
		b.Move(w, o, phys.CollisionBordersVector(w, phys.Vel()))
	}
}

// Move moves the object
func (b *ManualBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}

// Draw draws any artifacts of the behavior
func (b *ManualBehavior) Draw(win *pixelgl.Window) {

}

// TargetSeekerBehavior moves in shortest path to the target
type TargetSeekerBehavior struct {
	DefaultBehavior
	target   Target
	qt       *quadtree.Tree
	path     quadtree.NodeList
	fullpath []pixel.Vec
	cost     int
	source   pixel.Vec
	// finder          graph.PathFinder // path finder function
	finder          quadtree.PathFinder // path finder function
	turnsAtLocation int                 // number of turns at current location
	targetsCaught   int64

	// TODO: Change this to be based on expected steps rather than wall time
	targetAcquireTime    time.Time     // when this target was acquired
	maxTargetAcquireTime time.Duration // max allowed time to get to the target
}

// NewTargetSeekerBehavior return a TargetSeekerBehavior
func NewTargetSeekerBehavior(f quadtree.PathFinder) *TargetSeekerBehavior {
	b := &TargetSeekerBehavior{
		finder:               f,
		maxTargetAcquireTime: time.Second * time.Duration(utils.RandomInt(10, 20)),
	}
	b.name = "target_seeker"
	b.description = "Travels in shortest path to target, if given, otherwise stands still."
	return b
}

func (b *TargetSeekerBehavior) RemainingTargetAcquireTime() time.Duration {
	return (b.maxTargetAcquireTime - time.Since(b.targetAcquireTime)).Round(time.Millisecond)
}

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
func (b *TargetSeekerBehavior) populateMoveGraph(w *World) *quadtree.Tree {
	// log.Printf("Populating move graph for %v", o.Name())

	// augmented fixtures, these are what we check collisions against
	// they are grown by 1/2 size of object on each side to account for movement
	fixtures := []pixel.Rect{}

	phys := b.parent.Phys()

	cobjects, _ := w.CollisionObjects() // must be all for now
	for _, other := range cobjects {
		if b.parent.ID() == other.ID() {
			continue
		}
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+phys.Location().W(),
			other.Phys().Location().H()+phys.Location().H())
		scaled := other.Phys().Location().Resized(c, size)
		fixtures = append(fixtures, scaled)
	}

	// add start and target to the quadtree
	s := phys.Location().Center()
	t := b.target.Location()
	start := pixel.R(s.X, s.Y, s.X, s.Y)
	target := pixel.R(t.X, t.Y, t.X, t.Y)

	// Use own quadtree
	/////////////////////////////////////
	fixtures = append(fixtures, start, target)

	// minimum size of rectangle side at which we stop splitting
	// based on the size of the target seeker
	minSize := math.Min(phys.Location().W(), phys.Location().H())

	// quadtree
	qtBounds := pixel.R(
		w.Ground.Phys().Location().Min.X+phys.Location().W()/2, w.Ground.Phys().Location().Max.Y+phys.Location().H()/2,
		w.X-phys.Location().W()/2, w.Y-phys.Location().H()/2)
	qt, err := quadtree.NewTree(qtBounds, fixtures, minSize)
	if err != nil {
		log.Fatalf("error creating quadtree: %v", err)
	}

	startNode, err := qt.Locate(start.Center())
	if err != nil {
		log.Fatalf("%v", err)
	}
	targetNode, err := qt.Locate(target.Center())
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
func (b *TargetSeekerBehavior) FindPath(start, target pixel.Vec) (quadtree.NodeList, int, error) {

	// log.Printf("looking for path from %v to %v", start, target)
	path, cost, err := b.finder(b.qt, start, target)
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
	// drawTree, colorTree, drawText, drawObjects := true, false, false, true
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
		quadtree.DrawPath(win, drawPath, pathColor)
	}
	// draw the full path
	// quadtree.DrawPath(win, b.fullpath, pathColor)
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
