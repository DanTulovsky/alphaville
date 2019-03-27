package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"github.com/faiface/pixel"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Behavior is the interface for all behaviors
type Behavior interface {
	Description() string
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

	switch {
	case phys.MovingDown():
		if phys.CollisionBelow(w) {
			b.avoidCollisionBelow(phys)
			return true
		}
	case phys.MovingUp():
		if phys.CollisionAbove(w) {
			b.avoidCollisionAbove(phys, w)
			return true
		}
	case phys.MovingRight():
		if phys.CollisionRight(w) {
			b.avoidCollisionRight(phys)
			return true
		}
	case phys.MovingLeft():
		if phys.CollisionLeft(w) {
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
		// b.ChangeHorizontalDirection(phys)
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

	if phys.Vel().X != 0 && phys.Vel().Y != 0 {
		// cannot currently move in both X and Y direction
		log.Fatalf("o:%+#v\nx: %v; y: %v\n", o, phys.Vel().X, phys.Vel().Y)
	}

	// TODO: refactor to use CollisionBorders() function
	switch {
	case phys.MovingLeft() && phys.Location().Min.X+phys.Vel().X <= 0:
		// left border
		phys.SetLocation(phys.Location().Moved(pixel.V(0-phys.Location().Min.X, 0)))
		b.ChangeHorizontalDirection(phys)

	case phys.MovingRight() && phys.Location().Max.X+phys.Vel().X >= w.X:
		// right border
		phys.SetLocation(phys.Location().Moved(pixel.V(w.X-phys.Location().Max.X, 0)))
		b.ChangeHorizontalDirection(phys)

	case phys.MovingDown() && phys.Location().Min.Y+phys.Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		phys.SetLocation(phys.Location().Moved(pixel.V(0, w.Ground.Phys().Location().Max.Y-phys.Location().Min.Y)))
		v := phys.Vel()
		v.Y = 0
		v.X = phys.PreviousVel().X
		phys.SetVel(v)

	case phys.MovingUp() && phys.Location().Max.Y+phys.Vel().Y >= w.Y && phys.Vel().Y > 0:
		// stop at ceiling if going up
		phys.SetLocation(phys.Location().Moved(pixel.V(0, w.Y-phys.Location().Max.Y)))
		v := phys.Vel()
		v.Y = 0
		phys.SetVel(v)
		phys.SetCurrentMass(o.Mass())

	default:
		newLocation := phys.Location().Moved(pixel.V(v.X, v.Y))
		phys.SetLocation(newLocation)
	}
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

	if !phys.HaveCollision(w) {
		b.Move(w, o, phys.CollisionBorders(w, phys.Vel()))
	}
}

// Move moves the object
func (b *ManualBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}

// TargetSeekerBehavior moves in shortest path to the target
type TargetSeekerBehavior struct {
	DefaultBehavior
	target    Target
	moveGraph *graph.Graph
	path      []*graph.Node
	cost      int
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

// allCollisionVerticies returns a list of all verticies of all collision object
func (b *TargetSeekerBehavior) allCollisionVerticies(w *World, o Object) []pixel.Vec {
	l := []pixel.Vec{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}
		scaleX := (other.Phys().Location().Max.X - other.Phys().Location().Min.X) / 2
		scaleY := (other.Phys().Location().Max.Y - other.Phys().Location().Min.Y) / 2
		for _, v := range utils.RectVerticiesScaled(other.Phys().Location(), scaleX, scaleY, w.X, w.Y) {
			l = append(l, v)
		}
	}
	return l
}

// allCollisionEdges returns a list of all edges of all collision object
func (b *TargetSeekerBehavior) allCollisionEdges(w *World, o Object) []graph.Edge {
	l := []graph.Edge{}

	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}
		scaleX := (other.Phys().Location().Max.X - other.Phys().Location().Min.X) / 2
		scaleY := (other.Phys().Location().Max.Y - other.Phys().Location().Min.Y) / 2
		v := utils.RectVerticiesScaled(other.Phys().Location(), scaleX, scaleY, w.X, w.Y)
		r := pixel.R(v[0].X, v[0].Y, v[2].X, v[2].Y)

		for _, v := range graph.RectEdges(r) {
			l = append(l, v)
		}
	}
	return l
}

// isVisbile returns true if v is visibile from p (no intersecting edges)
func (b *TargetSeekerBehavior) isVisbile(w *World, p, v pixel.Vec, edges []graph.Edge) bool {
	log.Printf("points: %v, %v", p, v)
	for _, e := range edges {
		log.Printf("checking edge: %v", e)
		if (e.A == p || e.B == p) && (e.A == v || e.B == v) {
			// point are on the same segment, so visible
			log.Println("same segment")
			return true
		}

		// exclude edges that include v
		if e.A == v || e.B == v {
			log.Println("  excluded")
			continue
		}
		if graph.EdgesIntersect(graph.Edge{A: p, B: v}, e) {
			log.Println("  intersect")
			return false
		}
	}
	log.Printf("  included")
	return true
}

func (b *TargetSeekerBehavior) populateVisibilityGraph(w *World, o Object) {
	log.Printf("Populating visibility graph for %v", o.Name())
	g := graph.NewGraph()
	verticies := b.allCollisionVerticies(w, o)
	edges := b.allCollisionEdges(w, o)

	// Add all verticies (except source) to the graph
	for _, v := range verticies {
		g.AddNode(graph.NewItemNode(v, 1))
	}

	// source node
	p := graph.NewItemNode(o.Phys().Location().Center(), 0)
	g.AddNode(p)

	// target node
	t := graph.NewItemNode(b.target.Location(), 1)
	g.AddNode(t)
	log.Printf("target: %v", t.Value().V)

	// populate visibility information for all nodes
	for _, n := range g.Nodes() {
		for _, other := range g.Nodes() {
			if n.Value().V == other.Value().V {
				continue
			}
			// check if  v is visible from p
			if b.isVisbile(w, n.Value().V, other.Value().V, edges) {
				g.AddEdge(n, other)
			}
		}
	}

	b.moveGraph = g
	log.Printf("%v", b.moveGraph)
}

// SetTarget sets the target
func (b *TargetSeekerBehavior) SetTarget(t Target) {
	b.target = t
}

// Target returns the current target
func (b *TargetSeekerBehavior) Target() Target {
	return b.target
}

// nextDirectionToTarget returns the next direction to travel to the target
// up, down, left, right
func (b *TargetSeekerBehavior) nextDirectionToTarget(w *World, o Object) string {
	if len(b.path) > 0 && o.Phys().Location().Contains(b.path[0].Value().V) {
		log.Printf("removing node: %v", b.path[0])
		b.path = append(b.path[:0], b.path[1:]...)
		log.Printf("path is now: %v", b.path)
	}

	if len(b.path) == 0 {
		// log.Printf("path ran out...")
		return ""
	}

	t := b.path[0].Value().V
	c := o.Phys().Location().Center()

	to := t.To(c)

	switch {
	case to.X < 0 && !o.Phys().Location().Contains(pixel.V(t.X, c.Y)):
		return "right"
	case to.X > 0 && !o.Phys().Location().Contains(pixel.V(t.X, c.Y)):
		return "left"
	case to.Y < 0 && !o.Phys().Location().Contains(pixel.V(c.X, t.Y)):
		return "up"
	case to.Y > 0 && !o.Phys().Location().Contains(pixel.V(c.X, t.Y)):
		return "down"
	}

	return ""
}

// isAtTarget returns true if any part of the object covers the target
func (b *TargetSeekerBehavior) isAtTarget(o Object) bool {
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

// Direction returns the velocity vector setting the correct direction to travel
func (b *TargetSeekerBehavior) Direction(w *World, o Object) pixel.Vec {

	// find direction to move and set x, y based on velocity
	d := b.nextDirectionToTarget(w, o)

	switch d {
	case "up":
		return pixel.V(0, 1)
	case "down":
		return pixel.V(0, -1)
	case "right":
		return pixel.V(1, 0)
	case "left":
		return pixel.V(-1, 0)
	}
	return pixel.V(0, 0) // default is not to move
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
	return path, cost, err
}

// Update implements the Behavior Update method
func (b *TargetSeekerBehavior) Update(w *World, o Object) {
	if b.target == nil {
		if t, err := b.pickNewTarget(w); err == nil {
			b.SetTarget(t)
			b.populateVisibilityGraph(w, o)

			b.path, b.cost, err = b.FindPath(o.Phys().Location().Center(), b.target.Location())
			if err != nil {
				log.Printf("error finding path: %v", err)
			}
		} else {
			log.Printf("error picking target: %v", err)
		}
		return
	}

	if b.isAtTarget(o) {
		log.Println("is at target")
		return
	}

	// if b.moveGraph == nil {
	// 	b.populateVisibilityGraph(w, o)
	// }

	// if b.path == nil {
	// 	path, cost, err := b.FindPath(o.Phys().Location().Center(), b.target.Location())
	// 	if err != nil {
	// 		log.Printf("error finding path: %v", err)
	// 	}

	// 	b.path = path
	// 	log.Printf("path (cost=%v): %v", cost, path)
	// }

	phys := o.NextPhys()

	d := b.Direction(w, o)
	phys.SetManualVelocity(d)
	// o.Phys().SetManualVelocity(d)

	// check collisions with objects
	if !phys.HaveCollision(w) {
		b.Move(w, o, phys.CollisionBorders(w, phys.Vel()))
	}
}

// Move moves the object
func (b *TargetSeekerBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}
