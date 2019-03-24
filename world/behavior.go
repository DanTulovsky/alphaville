package world

import (
	"bytes"
	"html/template"
	"log"

	"golang.org/x/image/colornames"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Behavior is the interface for all behaviors
type Behavior interface {
	Description() string
	Draw(*pixelgl.Window) // draws any artifacts due to the behavior
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

// Draw does nothing
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

	if !phys.HaveCollision(w) {
		b.Move(w, o, phys.CollisionBorders(w, phys.Vel()))
	}
}

// Move moves the object
func (b *ManualBehavior) Move(w *World, o Object, v pixel.Vec) {
	newLocation := o.NextPhys().Location().Moved(pixel.V(v.X, v.Y))
	o.NextPhys().SetLocation(newLocation)
}

// Draw does nothing
func (b *ManualBehavior) Draw(win *pixelgl.Window) {

}

// TargetSeekerBehavior moves in shortest path to the target
type TargetSeekerBehavior struct {
	DefaultBehavior
	target pixel.Vec
}

// NewTargetSeekerBehavior return a TargetSeekerBehavior
func NewTargetSeekerBehavior() *TargetSeekerBehavior {
	b := &TargetSeekerBehavior{}
	b.name = "target_seeker"
	b.description = "Travels in shortest path to target, if given, otherwise stands still."
	return b
}

// SetTarget sets the target
func (b *TargetSeekerBehavior) SetTarget(t pixel.Vec) {
	b.target = t
}

// Target returns the current target
func (b *TargetSeekerBehavior) Target() pixel.Vec {
	return b.target
}

// nextDirectionToTarget returns the next direction to travel to the target
// up, down, left, right
func (b *TargetSeekerBehavior) nextDirectionToTarget(w *World, o Object) string {
	t := b.Target()
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
	return o.Phys().Location().Contains(b.Target())
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

// pickNewTarget sets a new random target
func (b *TargetSeekerBehavior) pickNewTarget(w *World) {
	target := pixel.V(utils.RandomFloat64(0, w.X), utils.RandomFloat64(0, w.Y))
	b.SetTarget(target)
}

// Draw draws the target
func (b *TargetSeekerBehavior) Draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)
	imd.Color = colornames.Red
	imd.Push(b.Target())
	imd.Circle(10, 0)
	imd.Draw(win)
}

// Update implements the Behavior Update method
func (b *TargetSeekerBehavior) Update(w *World, o Object) {
	if b.isAtTarget(o) {
		b.pickNewTarget(w)
		return
	}

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
