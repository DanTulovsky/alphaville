package world

import (
	"bytes"
	"html/template"
	"log"

	behave "github.com/askft/go-behave"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// DefaultBehavior is the default implementation of Behavior
type DefaultBehavior struct {
	description string
	name        string
	parent      Object
	t           *behave.BehaviorTree
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
	if b.ChangeVerticalDirection(w, o) {
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

// ChangeVerticalDirection updates the vertical direction if needed
func (b *DefaultBehavior) ChangeVerticalDirection(w *World, o Object) bool {
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

// Tree returns the behavior tree of this behavior
func (b *DefaultBehavior) Tree() *behave.BehaviorTree {
	return b.t

}
