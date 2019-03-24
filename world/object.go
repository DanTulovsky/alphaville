package world

import (
	"fmt"
	"image/color"
	"log"

	"golang.org/x/image/colornames"

	"gogs.wetsnow.com/dant/alphaville/utils"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
)

// Object is an object in the world
type Object interface {
	Behavior() Behavior
	BoundingBox(pixel.Vec) pixel.Rect
	CheckIntersect(*World)
	Draw(*pixelgl.Window)
	ID() uuid.UUID
	IsSpawned() bool
	Mass() float64
	NextPhys() ObjectPhys // returns the NextPhys object
	Name() string
	Phys() ObjectPhys // returns the Phys object
	Speed() float64
	SwapNextState()
	Update(*World) // Updates the object for the next iteration

	SetManualVelocity(v pixel.Vec)
	SetNextPhys(ObjectPhys)
	SetPhys(ObjectPhys)
}

// BaseObject is the base object
type BaseObject struct {
	name  string
	id    uuid.UUID
	color color.Color

	// initial Speed and Mass of BaseObject
	speed float64 // horizontal Speed (negative means move left)
	mass  float64

	// draws the BaseObject
	imd *imdraw.IMDraw

	// initial location of the BaseObject (bottom left corner)
	IX, IY float64

	// physics properties of the BaseObject
	phys     ObjectPhys
	nextPhys ObjectPhys // State of the object in the next round

	// behavior of the object (movement, etc...)
	behavior Behavior
}

// NewBaseObject return a new rectangular object
// phys bounding box is set based on width, height, unless phys is provided
func NewBaseObject(name string, color color.Color, speed, mass float64, behavior Behavior) BaseObject {
	o := BaseObject{}

	if behavior == nil {
		behavior = NewDefaultBehavior()
	}

	o.name = name
	o.id = uuid.New()
	o.color = color
	o.speed = speed
	o.mass = mass
	o.imd = imdraw.New(nil)
	o.phys = nil
	o.behavior = behavior

	return o
}

// BoundingBox must be implemented by each concrete object type; returns the bounding box of the object
func (o *BaseObject) BoundingBox(v pixel.Vec) pixel.Rect {
	log.Fatalf("using BaseObject BoundingBox, please implement: \n%#+v", o)
	return pixel.R(0, 0, 0, 0)
}

// Name returns the object's name
func (o *BaseObject) Name() string {
	return o.name
}

// Speed returns the object's speed
func (o *BaseObject) Speed() float64 {
	return o.speed
}

// Mass returns the object's mass
func (o *BaseObject) Mass() float64 {
	return o.mass
}

// ID returns the object's ID
func (o *BaseObject) ID() uuid.UUID {
	return o.id
}

// Behavior return the behavior object
func (o *BaseObject) Behavior() Behavior {
	return o.behavior
}

// SetBehavior sets the bahavior of an object
func (o *BaseObject) SetBehavior(b Behavior) {
	o.behavior = b
}

// Phys return the phys object
func (o *BaseObject) Phys() ObjectPhys {
	return o.phys
}

// NextPhys return the nextPhys object
func (o *BaseObject) NextPhys() ObjectPhys {
	return o.nextPhys
}

// SetPhys sets the phys object
func (o *BaseObject) SetPhys(op ObjectPhys) {
	o.phys = op
}

// SetNextPhys sets the nextPhys object
func (o *BaseObject) SetNextPhys(op ObjectPhys) {
	o.nextPhys = op
}

// IsSpawned returns true if the object already spawned in the world
func (o *BaseObject) IsSpawned() bool {
	return o.Phys() != nil
}

// Update the Object every frame
func (o *BaseObject) Update(w *World) {
	o.Behavior().Update(w, o)
	o.CheckIntersect(w)
}

// SwapNextState swaps the current state for next state of the object
func (o *BaseObject) SwapNextState() {
	if o.IsSpawned() {
		o.phys = o.nextPhys.Copy()
	}
}

// CheckIntersect prints out an error if this object intersects with another one
func (o *BaseObject) CheckIntersect(w *World) {
	for _, other := range w.CollisionObjects() {
		if o.ID() == other.ID() {
			continue // skip yourself
		}
		if o.NextPhys().Location().Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			log.Printf("%#+v (%v) intersects with %#v (%v)", o.name, o.NextPhys(), other.Name(), other.Phys())
		}
	}
}

// CheckIntersectRect prints out an error if this object intersects with another one
func CheckIntersectRect(w *World, r pixel.Rect, id uuid.UUID) {
	for _, other := range w.CollisionObjects() {
		if id == other.ID() {
			continue // skip yourself
		}
		if r.Intersect(other.Phys().Location()) != pixel.R(0, 0, 0, 0) {
			log.Printf("> %#+v (%v) intersects with %#v (%v)", id, r, other.Name(), other.Phys())

		}
	}
}

// Draw must be implemented by concrete objects
func (o *BaseObject) Draw(win *pixelgl.Window) {

	// draw name of the object
	txt := text.New(pixel.V(o.Phys().Location().Center().XY()), utils.Atlas())
	txt.Color = colornames.Red
	fmt.Fprintf(txt, "IMPLEMENT ME!")
	txt.Draw(win, pixel.IM)
}

// SetManualVelocity sets the velocity of the manually controlled object
func (o *BaseObject) SetManualVelocity(v pixel.Vec) {
	o.NextPhys().SetManualVelocity(v)
	o.Phys().SetManualVelocity(v)
}
