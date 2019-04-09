package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
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
	Update(*World, Object)
}

// DefaultBehavior is the default implementation of Behavior
type DefaultBehavior struct {
	description string
	name        string
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

	clocation := phys.HaveCollisionAt(w)

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

	if phys.HaveCollisionAt(w) == "" {
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
	target    Target
	moveGraph *graph.Graph
	qt        *quadtree.Tree
	path      []*graph.Node
	fullpath  []*graph.Node
	cost      int
	source    pixel.Vec
	finder    graph.PathFinder // path finder function
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

// scaledCollisionVerticies returns a list of all verticies of all collision objects
// scaled by half the size of o
func (b *TargetSeekerBehavior) scaledCollisionVerticies(w *World, o Object) []vertecy {
	v := []vertecy{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}

		// until movemement is fixed, add an additional buffer around object
		var buffer float64 = 6
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

		// until movemement is fixed, add an additional buffer around object
		var buffer float64 = 4
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+o.Phys().Location().W()+buffer,
			other.Phys().Location().H()+o.Phys().Location().H()+buffer)
		scaled := other.Phys().Location().Resized(c, size)

		edges := scaled.Edges()
		l = append(l, edges[0], edges[1], edges[2], edges[3])

	}
	return l
}

// isVisbile returns true if v is visibile from p (no intersecting edges)
func (b *TargetSeekerBehavior) isVisbile(w *World, p, v pixel.Vec, edges []pixel.Line, n, other *graph.Node) bool {
	for _, e := range edges {
		if (e.A == p && e.B == v) || (e.A == v && e.B == p) {
			// point are on the same segment, so visible
			return true
		}

		// exclude points on the same object if they are not on the same edge (we only deal with Rectangles here)
		if n.Object() == other.Object() {
			if p.X != v.X && p.Y != v.Y {
				return false
			}
		}
		// exclude edges that include v or p
		if e.A == v || e.B == v || e.A == p || e.B == p {
			continue
		}

		// Currently broken https://github.com/faiface/pixel/issues/175
		// once resolved, replace the code below
		// if _, isect := pixel.L(p, v).Intersect(e); isect {
		// 	return false
		// }

		ge := graph.Edge{A: p, B: v}
		etemp := graph.Edge{A: e.A, B: e.B}
		if graph.EdgesIntersect(ge, etemp) {
			return false
		}
	}
	return true
}

// populateVisibilityGraph creates a visibility graph
// the nodes are verticies of augmented rectangles
// the edges are paths between nodes that have no other node in between
// polygonal map from here:
// http://theory.stanford.edu/~amitp/GameProgramming/MapRepresentations.html#polygonal-maps
func (b *TargetSeekerBehavior) populateVisibilityGraph(w *World, o Object) {
	log.Printf("Populating visibility graph for %v", o.Name())
	g := graph.New()
	verticies := b.scaledCollisionVerticies(w, o)
	edges := b.scaledCollisionEdges(w, o)

	// Add all verticies (except source) to the graph
	for _, v := range verticies {
		g.AddNode(graph.NewItemNode(v.O, v.V, 1))
	}

	// source node
	p := graph.NewItemNode(o.ID(), o.Phys().Location().Center(), 0)
	g.AddNode(p)

	// target, not part of any object
	t := graph.NewItemNode(b.target.ID(), b.target.Location(), 1)
	g.AddNode(t)
	// log.Printf("target: %v", t.Value().V)

	// populate visibility information for all nodes
	for _, n := range g.Nodes() {
		// log.Printf(">> checking visibility from %v", n)
		for _, other := range g.Nodes() {
			if n.Value().V == other.Value().V {
				continue
			}
			// check if  v is visible from p
			if b.isVisbile(w, n.Value().V, other.Value().V, edges, n, other) {
				g.AddEdge(n, other)
			}
		}
	}

	b.moveGraph = g
	log.Printf("%v", b.moveGraph)
}

// populateMoveGraph creates a move graph by doing a
// cell decomposition.  the nodes are cells between the fixtures, and the edges are
// connections between them; from:
// https://cs.stanford.edu/people/eroberts/courses/soco/projects/1998-99/robotics/basicmotion.html
// https://www.dis.uniroma1.it/~oriolo/amr/slides/MotionPlanning1_Slides.pdf
// o is the target seeker
func (b *TargetSeekerBehavior) populateMoveGraph(w *World, o Object) {
	log.Printf("Populating move graph for %v", o.Name())

	// augmented fixtures, these are what we check collisions against
	// they are grown by 1/2 size of object on each side to account for movement
	fixtures := []pixel.Rect{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue
		}
		var buffer float64 = 2
		c := other.Phys().Location().Center()
		size := pixel.V(other.Phys().Location().W()+o.Phys().Location().W()+buffer,
			other.Phys().Location().H()+o.Phys().Location().H()+buffer)
		scaled := other.Phys().Location().Resized(c, size)
		fixtures = append(fixtures, scaled)
	}

	// add start and target to the quadtree
	s := o.Phys().Location().Center()
	t := b.target.Location()
	start := pixel.R(s.X, s.Y, s.X+1, s.Y+1)
	target := pixel.R(t.X, t.Y, t.X+1, t.Y+1)
	fixtures = append(fixtures, target)

	// minimum size of rectangle side at which we stop splitting
	minSize := float64(6)

	// quadtree
	qt, err := quadtree.NewTree(pixel.R(w.Ground.Phys().Location().Min.X, w.Ground.Phys().Location().Max.Y, w.X, w.Y), fixtures, minSize)
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

// isAtTarget returns true if any part of the object covers the target
func (b *TargetSeekerBehavior) isAtTarget(o Object) bool {

	if o.Phys().Location().Intersect(b.target.Bounds()) != pixel.R(0, 0, 0, 0) {

		o.Notify(NewObjectEvent(
			fmt.Sprintf("[%v] found target [%v]", o.Name(), b.target.Name()), time.Now(),
			observer.EventData{Key: "target_found", Value: b.target.Name()}))
		b.target.Destroy()
		b.target = nil

		return true
	}
	return false
}

// Direction returns the next direction to travel to the target
func (b *TargetSeekerBehavior) Direction(w *World, o Object) pixel.Vec {
	// remove the current location from path
	circle := pixel.C(o.Phys().Location().Center(), 2)
	if len(b.path) > 0 && circle.Contains(b.path[0].Value().V) {
		// if len(b.path) > 0 && o.Phys().Location().Contains(b.path[0].Value().V) {

		b.source = b.path[0].Value().V
		b.path = append(b.path[:0], b.path[1:]...)
	}

	if len(b.path) == 0 {
		// log.Printf("path ran out...")
		return pixel.ZV
	}
	source := b.source
	// target is the next node in the path
	target := b.path[0].Value().V
	// current location of target seeker
	c := o.Phys().Location().Center()

	log.Printf("From: %v; To: %v\n", source, target)
	log.Printf("  Current location: %v", o.Phys().Location().Center())
	var moves []pixel.Vec

	orient := graph.Orientation(source, target, c)
	log.Printf("  orientation: %v", orient)

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
		log.Println(moves)
		return moves[utils.RandomInt(0, len(moves))]
	}

	o.SetManualVelocity(pixel.ZV)
	return pixel.ZV

}

// pickNewTarget sets a new random target if available
func (b *TargetSeekerBehavior) pickNewTarget(w *World) (Target, error) {
	log.Println("Picking new target...")
	targets := w.AvailableTargets()
	if len(targets) == 0 {
		return nil, fmt.Errorf("no available targets")
	}

	t := targets[utils.RandomInt(0, len(targets))]
	log.Printf("Picked new target %v", t.Location())
	return t, nil
}

// FindPath returns the path and cost between start and target
func (b *TargetSeekerBehavior) FindPath(start, target pixel.Vec) ([]*graph.Node, int, error) {

	log.Printf("looking for path from %v to %v", start, target)
	path, cost, err := b.finder(b.moveGraph, start, target)
	if err != nil {
		log.Printf("error finding path: %v", err)
		return nil, 0, err
	}

	// add the path from the center of the quadrant to the target inside of it
	path = append(path, graph.NewItemNode(uuid.New(), b.target.Location(), 0))

	return path, cost, err
}

// Update implements the Behavior Update method
func (b *TargetSeekerBehavior) Update(w *World, o Object) {
	if b.target == nil {
		if t, err := b.pickNewTarget(w); err == nil {
			b.SetTarget(t)
			// b.populateVisibilityGraph(w, o)
			b.populateMoveGraph(w, o)

			log.Printf("qt: %v", b.qt)
			log.Printf("g: %v", b.moveGraph)
			startNode := b.qt.Locate(o.Phys().Location().Center())
			targetNode := b.qt.Locate(b.target.Bounds().Center())

			log.Printf("startNode: %v (o at: %v)", startNode.Bounds(), o.Phys().Location().Center())
			log.Printf("targetNode: %v", targetNode.Bounds())

			b.path, b.cost, err = b.FindPath(startNode.Bounds().Center(), targetNode.Bounds().Center())
			if err != nil {
				log.Printf("error finding path: %v", err)
			}

			b.fullpath = []*graph.Node{}
			for _, n := range b.path {
				b.fullpath = append(b.fullpath, n)
			}
			sn := graph.NewItemNode(uuid.New(), o.Phys().Location().Center(), 0)
			b.fullpath = append([]*graph.Node{sn}, b.fullpath...)
			b.source = o.Phys().Location().Center()
			log.Printf("Path found: %v", b.fullpath)

		} else {
			log.Printf("error picking target: %v", err)
		}
		return
	}

	if b.isAtTarget(o) {
		log.Println("is at target")
		return
	}

	phys := o.NextPhys()

	d := b.Direction(w, o)
	phys.SetManualVelocity(d)
	// o.Phys().SetManualVelocity(d)

	if phys.HaveCollisionAt(w) == "" {
		// move, checking collisions with world borders
		b.Move(w, o, phys.CollisionBordersVector(w, phys.Vel()))
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
	drawTree, drawText, drawObjects := true, false, true
	b.qt.Draw(win, drawTree, drawText, drawObjects)

	// draw the path
	graph.DrawPath(win, b.fullpath)

	// // Draw the path
	// imd.Color = colornames.Lightblue
	// // Draw the graph lines
	// for n, other := range b.moveGraph.Edges() {
	// 	for _, o := range other {
	// 		imd.Push(n.Value().V)
	// 		imd.Push(o.Value().V)
	// 		imd.Line(1)
	// 	}
	// }

	// // draw the graph
	// imd.Color = b.target.Color()
	// for _, p := range b.fullpath {
	// 	imd.Push(p.Value().V)
	// }
	// imd.Line(1)
	// imd.Draw(win)

}
