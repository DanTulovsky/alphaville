package world

import "github.com/faiface/pixel"

// Behavior is the interface for all behaviors
type Behavior interface {
	Update(*World, Object)
}

// DefaultBehavior is the default implementation of Behavior
type DefaultBehavior struct {
}

// NewDefaultBehavior return a DefaultBehavior
func NewDefaultBehavior() *DefaultBehavior {

	return &DefaultBehavior{}
}

// Update executes the next world step for the object
// It updates the NextPhys() of the object for next step based on the encoded behavior
func (b *DefaultBehavior) Update(w *World, o Object) {

	// Movement and location are set in the NextPhys object
	phys := o.NextPhys()

	// if on the ground and X velocity is 0, reset it - this seems to be a bug
	if phys.OnGround(w) && phys.Stopped() {
		v := o.NextPhys().Vel()
		v.X = o.NextPhys().PreviousVel().X
		if v.X == 0 {
			v.X = o.Speed()
		}
		v.Y = 0
		o.NextPhys().SetVel(v)
	}

	// check if object should rise or fall, these checks not based on collisions
	phys.ChangeVerticalDirection(w)

	// check collisions and adjust movement parameters
	// if a collision is detected, no movement happens this round
	if phys.HandleCollisions(w) {
		return
	}

	// no collisions detected, move
	phys.Move(w, pixel.V(o.NextPhys().Vel().X, o.NextPhys().Vel().Y))
}
