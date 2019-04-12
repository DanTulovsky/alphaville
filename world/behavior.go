package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"math"
	"time"

	"gogs.wetsnow.com/dant/alphaville/quadtree"

	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"

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
	target          Target
	moveGraph       *graph.Graph
	qt              *quadtree.Tree
	path            []*graph.Node
	fullpath        []*graph.Node
	cost            int
	source          pixel.Vec
	finder          graph.PathFinder // path finder function
	turnsAtLocation int              // number of turns at current location
	targetsCaught   int64
}

// NewTargetSeekerBehavior return a TargetSeekerBehavior
func NewTargetSeekerBehavior(f graph.PathFinder) *TargetSeekerBehavior {
	b := &TargetSeekerBehavior{
		moveGraph: nil,
		finder:    f,
	}
	b.name = "target_seeker"
	b.description = "Travels in shortest path to target, if given, otherwise stands still."
	return b
}

type vertecy struct {
	V pixel.Vec
	O uuid.UUID
}

// String returns ...
func (b *TargetSeekerBehavior) String() string {
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("physObject").Parse(
		`
Behavior
  Name: {{.Name}}	
	Desc: {{.Description}}	
	Target: {{.Target.Location}}
	Turns At Location: {{.TurnsBlocked}}
	Targets Caught: {{.TargetsCaught}}
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

// scaledCollisionVerticies returns a list of all verticies of all collision objects
// scaled by half the size of o
func (b *TargetSeekerBehavior) scaledCollisionVerticies(w *World, o Object) []vertecy {
	v := []vertecy{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}

		// until movemment is fixed, add an additional buffer around object
		var buffer float64 = 0
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+o.Phys().Location().W()+buffer,
			other.Phys().Location().H()+o.Phys().Location().H()+buffer)
		vertecies := other.Phys().Location().Resized(c, size).Vertices()

		for _, vr := range vertecies {
			v = append(v, vertecy{V: vr, O: other.ID()})
		}
	}
	return v
}

// scaledCollisionEdges returns a list of all edges of all collision objects
// scaled by half the size of o
func (b *TargetSeekerBehavior) scaledCollisionEdges(w *World, o Object) []pixel.Line {
	l := []pixel.Line{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}

		// until movement is fixed, add an additional buffer around object
		var buffer float64 = 0
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+o.Phys().Location().W()+buffer,
			other.Phys().Location().H()+o.Phys().Location().H()+buffer)
		scaled := other.Phys().Location().Resized(c, size)

		edges := scaled.Edges()
		l = append(l, edges[0], edges[1], edges[2], edges[3])

	}
	return l
}

// populateMoveGraph creates a move graph by doing a
// cell decomposition.  the nodes are cells between the fixtures, and the edges are
// connections between them; from:
// https://cs.stanford.edu/people/eroberts/courses/soco/projects/1998-99/robotics/basicmotion.html
// https://www.dis.uniroma1.it/~oriolo/amr/slides/MotionPlanning1_Slides.pdf
// o is the target seeker
func (b *TargetSeekerBehavior) populateMoveGraph(w *World) {
	// log.Printf("Populating move graph for %v", o.Name())

	// augmented fixtures, these are what we check collisions against
	// they are grown by 1/2 size of object on each side to account for movement
	fixtures := []pixel.Rect{}

	phys := b.parent.Phys()

	for _, other := range w.CollisionObjects() {
		if b.parent.ID() == other.ID() {
			continue
		}
		// TODO remove this buffer
		var buffer float64 = 1 // must be larger than quadtree.NewTree( ... minSize), why?
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+phys.Location().W()+buffer,
			other.Phys().Location().H()+phys.Location().H()+buffer)
		scaled := other.Phys().Location().Resized(c, size)
		fixtures = append(fixtures, scaled)
	}

	// add start and target to the quadtree
	s := phys.Location().Center()
	t := b.target.Location()
	start := pixel.R(s.X, s.Y, s.X+1, s.Y+1)
	target := pixel.R(t.X, t.Y, t.X+1, t.Y+1)
	fixtures = append(fixtures, target)

	// minimum size of rectangle side at which we stop splitting
	// based on the size of the target seeker
	minSize := math.Min(phys.Location().W(), phys.Location().H())

	// quadtree
	qtBounds := pixel.R(
		w.Ground.Phys().Location().Min.X+phys.Location().W(), w.Ground.Phys().Location().Max.Y+phys.Location().H(),
		w.X-phys.Location().W(), w.Y-phys.Location().H())
	qt, err := quadtree.NewTree(qtBounds, fixtures, minSize)
	if err != nil {
		log.Fatalf("error creating quadtree: %v", err)
	}

	b.qt = qt
	b.moveGraph = qt.ToGraph(start, target)
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

	if utils.VecLen(o.Phys().Location().Center(), b.target.Bounds().Center()) < o.Speed() {

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
	for len(b.path) > 0 && utils.VecLen(o.Phys().Location().Center(), b.path[0].Value().V) < o.Speed() {
		// if len(b.path) > 0 && o.Phys().Location().Contains(b.path[0].Value().V) {

		b.source = b.path[0].Value().V
		b.path = append(b.path[:0], b.path[1:]...)
	}

	if len(b.path) == 0 {
		// log.Printf("path ran out...")
		return pixel.ZV, pixel.V(0, 0)
	}
	source := b.source
	// target is the next node in the path
	target := b.path[0].Value().V
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
func (b *TargetSeekerBehavior) FindPath(start, target pixel.Vec) ([]*graph.Node, int, error) {

	// log.Printf("looking for path from %v to %v", start, target)
	path, cost, err := b.finder(b.moveGraph, start, target)
	if err != nil {
		return nil, 0, err
	}

	// add the path from the center of the quadrant to the target inside of it
	path = append(path, graph.NewItemNode(uuid.New(), b.target.Location(), 0))

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
	b.recalculateMoveInfo(w, o)

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
				log.Printf("[%v] target [%v] destroyed need to pick another one", b.parent.Name(), data.Value)
				b.target.Deregister(b)
				b.target = nil
			}
		}
	}
}

// OnNotify runs when a notification is received
func (b *TargetSeekerBehavior) OnNotify(e observer.Event) {
	switch event := e.(type) {
	case nil:
		log.Printf("nil notification")
	case *TargetEvent:
		b.processTargetEvent(event)
	}
}

// recalculateMoveInfo recalculates the path for an existing target
func (b *TargetSeekerBehavior) recalculateMoveInfo(w *World, o Object) {
	phys := o.NextPhys()

	b.populateMoveGraph(w)
	var err error
	startNode, err := b.qt.Locate(phys.Location().Center())
	if err != nil {
		log.Fatal(err)
	}
	targetNode, err := b.qt.Locate(b.target.Bounds().Center())
	if err != nil {
		log.Fatal(err)
	}

	b.path, b.cost, err = b.FindPath(startNode.Bounds().Center(), targetNode.Bounds().Center())
	if err != nil {
		// log.Printf("error finding path: %v", err)
	}

	b.fullpath = []*graph.Node{}
	for _, n := range b.path {
		b.fullpath = append(b.fullpath, n)
	}
	sn := graph.NewItemNode(uuid.New(), phys.Location().Center(), 0)
	b.fullpath = append([]*graph.Node{sn}, b.fullpath...)
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
		}
		return
	}

	if b.isAtTarget(o) {
		b.targetsCaught++
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
	drawTree, colorTree, drawText, drawObjects := true, false, false, true
	b.qt.Draw(win, drawTree, colorTree, drawText, drawObjects)

	pathColor := b.parent.Color()
	// draw the path from current location
	// l := graph.NewItemNode(uuid.New(), b.parent.Phys().Location().Center(), 0)
	// graph.DrawPath(win, append([]*graph.Node{l}, b.path...), pathColor)

	// draw the path full path
	graph.DrawPath(win, b.fullpath, pathColor)
}
