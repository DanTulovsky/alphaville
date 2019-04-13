package world

import (
	"bytes"
	"html/template"
	"log"

	"github.com/faiface/pixel"
)

// ObjectPhys is the physics of an object, these values change as the object moves
type ObjectPhys interface {
	Copy() ObjectPhys

	CollisionBordersVector(*World, pixel.Vec) pixel.Vec
	CurrentMass() float64
	HaveCollisionsAt(*World) []string
	IsAboveGround(w *World) bool
	IsZeroMass() bool
	Location() pixel.Rect
	LocationOf(Object) string
	OnGround(*World) bool
	MovingUp() bool
	MovingDown() bool
	MovingLeft() bool
	MovingRight() bool
	ParentObject() Object
	PreviousVel() pixel.Vec
	SetManualVelocity(v pixel.Vec)
	SetManualVelocityXY(v pixel.Vec)
	// SetManualVelocityXY(v pixel.Vec)
	Stop()
	Stopped() bool
	StoppedX() bool
	StoppedY() bool
	Vel() pixel.Vec

	SetCurrentMass(float64)
	SetLocation(pixel.Rect)
	SetPreviousVel(pixel.Vec)
	SetVel(pixel.Vec)
}

// BaseObjectPhys defines the physical (dynamic) object properties
type BaseObjectPhys struct {

	// current horizontal and vertical Speed of Object
	vel pixel.Vec

	// previous horizontal and vertical Speed of Object
	previousVel pixel.Vec

	// currentMass of the Object
	currentMass float64

	// this is the bounding rectangle in the world
	rect pixel.Rect

	parentObject Object
}

// String returns the Phys object string representation
func (o *BaseObjectPhys) String() string {
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("physObject").Parse(
		`
Phys
  Vel: {{.Vel}}	
  PreviousVel: {{.PreviousVel}}	
  CurrentMass: {{.CurrentMass}}	
  Rect: {{.Location}}
  ParentObject: {{.ParentObject.Name}}
`)

	if err != nil {
		log.Fatalf("object conversion error: %v", err)
	}
	err = tmpl.Execute(buf, o)
	if err != nil {
		log.Fatalf("object conversion error: %v", err)
	}

	return buf.String()
}

// ParentObject returns the parent object
func (o *BaseObjectPhys) ParentObject() Object {
	return o.parentObject
}

// SetParentObject sets the parent object
func (o *BaseObjectPhys) SetParentObject(po Object) {
	o.parentObject = po
}

// CurrentMass returns the current mass
func (o *BaseObjectPhys) CurrentMass() float64 {
	return o.currentMass
}

// PreviousVel returns the previous velocity vector
func (o *BaseObjectPhys) PreviousVel() pixel.Vec {
	return o.previousVel
}

// Vel returns the current velocity vector
func (o *BaseObjectPhys) Vel() pixel.Vec {
	return o.vel
}

// SetCurrentMass sets the current mass
func (o *BaseObjectPhys) SetCurrentMass(m float64) {
	o.currentMass = m
}

// SetPreviousVel sets the previous velocity vector
func (o *BaseObjectPhys) SetPreviousVel(v pixel.Vec) {
	o.previousVel = v
}

// SetVel sets the current velocity vector
func (o *BaseObjectPhys) SetVel(v pixel.Vec) {
	o.vel = v
}

// NewBaseObjectPhys return a new physic object
func NewBaseObjectPhys(rect pixel.Rect, po Object) ObjectPhys {

	return &BaseObjectPhys{
		rect:         rect,
		parentObject: po, // the object this ObjectPhys belongs to
	}
}

// Copy return a new rectangle phys object based on an existing one
func (o *BaseObjectPhys) Copy() ObjectPhys {

	op := NewBaseObjectPhys(o.Location(), o.parentObject)
	op.SetVel(pixel.V(o.Vel().X, o.Vel().Y))
	op.SetPreviousVel(pixel.V(o.PreviousVel().X, o.PreviousVel().Y))
	op.SetCurrentMass(o.CurrentMass())
	return op
}

// Location returns the current location
func (o *BaseObjectPhys) Location() pixel.Rect {
	return o.rect
}

// SetLocation sets the current location
func (o *BaseObjectPhys) SetLocation(r pixel.Rect) {
	o.rect = r
}

// OnGround returns true if object is on the ground
func (o *BaseObjectPhys) OnGround(w *World) bool {
	return o.Location().Min.Y == w.Ground.Phys().Location().Max.Y
}

// Stopped returns true if object is stopped in both directions
func (o *BaseObjectPhys) Stopped() bool {
	return o.Vel().X == 0 && o.Vel().Y == 0
}

// StoppedX returns true if object is stopped horizontally
func (o *BaseObjectPhys) StoppedX() bool {
	return o.Vel().X == 0
}

// StoppedY returns true if object is stopped vertically
func (o *BaseObjectPhys) StoppedY() bool {
	return o.Vel().Y == 0
}

// IsAboveGround checks if object is above ground
func (o *BaseObjectPhys) IsAboveGround(w *World) bool {
	return o.Location().Min.Y > w.Ground.Phys().Location().Max.Y
}

// IsZeroMass checks if object has no mass
func (o *BaseObjectPhys) IsZeroMass() bool {
	return o.CurrentMass() == 0
}

// MovingUp returns true if object is moving up
func (o *BaseObjectPhys) MovingUp() bool {
	return o.Vel().Y > 0
}

// MovingDown returns true if object is moving down
func (o *BaseObjectPhys) MovingDown() bool {
	return o.Vel().Y < 0
}

// MovingLeft returns true if object is moving left
func (o *BaseObjectPhys) MovingLeft() bool {
	return o.Vel().X < 0
}

// MovingRight returns true if object is moving right
func (o *BaseObjectPhys) MovingRight() bool {
	return o.Vel().X > 0
}

// LocationOf returns the location of other (above, below, left, right)
func (o *BaseObjectPhys) LocationOf(other Object) string {

	oMin := o.Location().Min
	oMax := o.Location().Max
	otherMin := other.Phys().Location().Min
	otherMax := other.Phys().Location().Max

	switch {
	case oMin.Y > otherMax.Y:
		return "below"
	case oMax.Y < otherMin.Y:
		return "above"
	case oMin.X > otherMax.X:
		return "left"
	case oMax.X < otherMin.X:
		return "right"
	}

	return ""
}

// HaveCollisionsAt returns all the location of the collisions (above, below, left, right) if there
// is a collision, otherwise ""
func (o *BaseObjectPhys) HaveCollisionsAt(w *World) []string {
	collisions := []string{}
	// collisionObjects, err := w.CollisionObjects()
	collisionObjects, err := w.CollisionObjectsWith(o.parentObject)
	if err != nil {
		log.Fatalf("%v is not in the world qt", o.parentObject.Name())
	}

	// Check collisions only with objects that intersect the same quadrant that fully contains o
	for _, other := range collisionObjects {
		log.Printf("Checking collisions with %v objects", len(collisionObjects))
		if o.parentObject.ID() == other.ID() {
			continue // skip yourself
		}

		// location of other compared to o (above, below, right, left)
		l := o.LocationOf(other)
		// log.Printf("o: %v; other: %v; l: %v", o.parentObject.Name(), other.Name(), l)
		// log.Printf("o is: movingRight? %v; movingLeft? %v; movingUp? %v; movingDown? %v", o.MovingRight(), o.MovingLeft(), o.MovingUp(), o.MovingDown())

		// log.Printf("other is: movingRight? %v; movingLeft? %v; movingUp? %v; movingDown? %v", other.Phys().MovingRight(), other.Phys().MovingLeft(), other.Phys().MovingUp(), other.Phys().MovingDown())

		switch l {
		case "left":
			if o.MovingRight() {
				continue
			}
		case "right":
			if o.MovingLeft() {
				continue
			}
		case "above":
			if o.MovingDown() {
				continue
			}
		case "below":
			if o.MovingUp() {
				continue
			}
		}

		// other moves as planned based on current velocity
		if HaveCollisions(o.Location(), other.Phys().Location(), o.Vel(), other.Phys().Vel()) {
			collisions = append(collisions, l)
		}
		// other doesn't move
		if HaveCollisions(o.Location(), other.Phys().Location(), o.Vel(), pixel.V(0, 0)) {
			collisions = append(collisions, l)
		}
	}

	return collisions
}

// CollisionBordersVector returns a movement vector that avoids collision with outside world border given vel vector
// If no collisions detected, vel is returned as is
func (o *BaseObjectPhys) CollisionBordersVector(w *World, vel pixel.Vec) pixel.Vec {

	switch {
	case o.MovingLeft() && o.Location().Min.X+vel.X <= 0:
		// left border
		return pixel.V(0-o.Location().Min.X, 0)
	case o.MovingRight() && o.Location().Max.X+vel.X >= w.X:
		// right border
		return pixel.V(w.X-o.Location().Max.X, 0)
	case o.MovingDown() && o.Location().Min.Y+o.Vel().Y < w.Ground.Phys().Location().Max.Y:
		// stop at ground level
		return pixel.V(0, w.Ground.Phys().Location().Max.Y-o.Location().Min.Y)
	case o.MovingUp() && o.Location().Max.Y+o.Vel().Y >= w.Y:
		// stop at ceiling if going up
		return pixel.V(0, w.Y-o.Location().Max.Y)
	}
	return vel
}

// Stop stops the object
func (o *BaseObjectPhys) Stop() {
	v := o.Vel()
	v.X = 0
	v.Y = 0
	o.SetVel(v)
}

// SetManualVelocity sets the manual object's velocity
func (o *BaseObjectPhys) SetManualVelocity(v pixel.Vec) {

	switch {
	case v.X < 0:
		v.X = o.ParentObject().Speed() * -1
		v.Y = 0
	case v.X > 0:
		v.X = o.ParentObject().Speed()
		v.Y = 0
	case v.Y > 0:
		v.Y = o.ParentObject().Speed()
		v.X = 0
	case v.Y < 0:
		v.Y = o.ParentObject().Speed() * -1
		v.X = 0
	}

	o.SetVel(v)
}

// SetManualVelocityXY sets the manual object's velocity
func (o *BaseObjectPhys) SetManualVelocityXY(v pixel.Vec) {

	switch {
	case v.X < 0:
		v.X = o.ParentObject().Speed() * -1
	case v.X > 0:
		v.X = o.ParentObject().Speed()
	}
	switch {
	case v.Y > 0:
		v.Y = o.ParentObject().Speed()
	case v.Y < 0:
		v.Y = o.ParentObject().Speed() * -1
	}

	o.SetVel(v)
}
