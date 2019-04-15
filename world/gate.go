package world

import (
	"fmt"
	"image/color"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/imdraw"

	"github.com/faiface/pixel"
)

// GateStatus is the status of the gate
type GateStatus int

const (
	// Unknown gate state
	Unknown = iota

	// GateOpen is open
	GateOpen

	// GateClosed is closed, nothing can come out of it
	GateClosed
)

// GateEvent implements the observer.Event interface to send events to other components
type GateEvent struct {
	observer.BaseEvent
}

// NewGateEvent create a new gate event
func NewGateEvent(d string, t time.Time, data ...observer.EventData) observer.Event {
	e := &GateEvent{}
	e.SetData(data)
	e.SetDescription(d)
	e.SetTime(t)

	return e
}

// Gate is a point in the world where new objects can appear
type Gate struct {
	id         uuid.UUID
	name       string
	Location   pixel.Vec
	Status     GateStatus
	Reserved   bool // gate is reserved by an object to be used next turn
	ReservedBy uuid.UUID

	// Wait this long before allowing a new spawn
	SpawnCoolDown time.Duration
	LastSpawn     time.Time

	Radius float64 // size

	// who to notify on events
	observers []observer.EventObserver

	// filters controls what object is allowed to use (reserver and spawn) this gate
	// if any filter denies (returns false), usage is not allowed
	filters []GateFilter
}

// GateFilter filters what object is allowed to use the gate, it should return true if allowed
type GateFilter func(Object) bool

// DefaultGateFilter allows all objects
var DefaultGateFilter = func(Object) bool {
	return true
}

// NewGate creates a new gate in the world
func NewGate(n string, l pixel.Vec, s GateStatus, coolDown time.Duration, radius float64, filters ...GateFilter) *Gate {

	g := &Gate{
		id:            uuid.New(),
		name:          n,
		Location:      l,
		Status:        s,
		Reserved:      false,
		SpawnCoolDown: coolDown,
		Radius:        radius,
		filters:       filters,
	}

	return g
}

// String returns the gate as string
func (g *Gate) String() string {
	return fmt.Sprintf("[%v] L: %v, S: %v, R: %v, C: %v (%v)", g.name, g.Location, g.Status, g.Reserved, g.SpawnCoolDown, g.LastSpawn)
}

// CanSpawn returns true if the gate can spawn
func (g *Gate) CanSpawn() bool {
	switch {
	case g.Status != GateOpen:
		return false
	case g.Reserved:
		return false
	case time.Since(g.LastSpawn) < g.SpawnCoolDown:
		return false
	}
	return true
}

// Reserve reserves a gate if it's available
func (g *Gate) Reserve(o Object) error {
	for _, f := range g.filters {
		if !f(o) {
			return fmt.Errorf("usage of gate denied by filter")
		}
	}

	if g.CanSpawn() {
		g.Reserved = true
		g.ReservedBy = o.ID()
		g.Notify(NewGateEvent(fmt.Sprintf("gate [%v] reserved for [%v]", g, o.ID()), time.Now()))

		return nil
	}
	return fmt.Errorf("gate %#v already reserved or closed", g)
}

// Release removes a gates reservation
func (g *Gate) Release() {
	g.Notify(NewGateEvent(fmt.Sprintf("gate [%v] reservation released", g), time.Now()))
	g.Reserved = false
	g.LastSpawn = time.Now()
}

// Draw draws the gate on the screen
func (g *Gate) Draw(win *pixelgl.Window) {
	imd := imdraw.New(nil)

	imd.Color = g.Color()

	imd.Push(g.Location)
	imd.Circle(g.Radius, 2)
	imd.Draw(win)

	// remaining time until next spawn
	txt := text.New(g.Location, utils.Atlas())
	txt.Color = colornames.Yellow

	label := ""

	switch {
	case g.Status == GateClosed:
		label = "inf"
	case !g.CanSpawn():
		label = fmt.Sprintf("%v", g.SpawnCoolDown-time.Since(g.LastSpawn).Truncate(time.Second))
	default:
		label = "inf"

	}

	// center the text
	txt.Dot.X -= txt.BoundsOf(label).W() / 2

	fmt.Fprintf(txt, label)
	txt.Draw(win, pixel.IM)
}

// Implement the observer.EventNotifier interface

// Register registers a new observer for notifying on.
func (g *Gate) Register(obs observer.EventObserver) {
	g.observers = append(g.observers, obs)
}

// Deregister de-registers an observer for notifying on.
func (g *Gate) Deregister(obs observer.EventObserver) {
	for i := 0; i < len(g.observers); i++ {
		if obs == g.observers[i] {
			g.observers = append(g.observers[:i], g.observers[i+1:]...)
		}
	}
}

// Notify notifies all observers on an event.
func (g *Gate) Notify(event observer.Event) {
	for i := 0; i < len(g.observers); i++ {
		g.observers[i].OnNotify(event)
	}
}

// Implement the Object interface

// Behavior returns nil
func (g *Gate) Behavior() Behavior {
	return nil
}

// BoundingBox returns the bonding box of the gate
func (g *Gate) BoundingBox(v pixel.Vec) pixel.Rect {
	min := pixel.V(g.Location.X-g.Radius, g.Location.Y-g.Radius)
	max := pixel.V(g.Location.X+g.Radius, g.Location.Y+g.Radius)

	return pixel.R(min.X, min.Y, max.X, max.Y)
}

// Size returns the size
func (g *Gate) Size() pixel.Rect {
	return g.BoundingBox(g.Location)
}

// ID returns the id
func (g *Gate) ID() uuid.UUID {
	return g.id
}

// IsSpawned always return false
func (g *Gate) IsSpawned() bool {
	return false
}

// Mass always returns -1
func (g *Gate) Mass() float64 {
	return -1
}

// NextPhys always returns nil
func (g *Gate) NextPhys() ObjectPhys {
	return nil
}

// Color returns the gate color
func (g *Gate) Color() color.Color {
	if g.Reserved || g.Status == GateClosed {
		return colornames.Red
	}
	return colornames.Green
}

// Name always returns 'null'
func (g *Gate) Name() string {
	return g.name
}

// Phys always returns nil
func (g *Gate) Phys() ObjectPhys {
	return nil
}

// Speed always returns 0
func (g *Gate) Speed() float64 {
	return 0
}

// SwapNextState does nothing
func (g *Gate) SwapNextState() {}

// Update does nothing
func (g *Gate) Update(*World) {}

// SetManualVelocity does nothing
func (g *Gate) SetManualVelocity(v pixel.Vec) {}

// SetNextPhys does nothing
func (g *Gate) SetNextPhys(ObjectPhys) {}

// SetPhys does nothing
func (g *Gate) SetPhys(ObjectPhys) {}

// CheckIntersect does nothing
func (g *Gate) CheckIntersect(*World) {}
