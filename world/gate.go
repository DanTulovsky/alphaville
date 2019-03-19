package world

import (
	"fmt"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/behavior"
	"golang.org/x/image/colornames"

	"github.com/faiface/pixel/imdraw"

	"github.com/faiface/pixel"
)

// gateStatus is the status of the gate
type gateStatus int

const (
	// Unknown gate state
	Unknown = iota

	// GateOpen is open
	GateOpen

	// GateClosed is closed, nothing can come out of it
	GateClosed
)

// GateEvent implements the behavior.Event interface to send events to other components
type GateEvent struct {
	description string
	time        time.Time // event time
}

// Description returns the event description
func (e *GateEvent) Description() string {
	return e.description
}

// String returns the event as string
func (e *GateEvent) String() string {
	return fmt.Sprintf("[%v] %v", e.time, e.description)
}

// Time returns the event time
func (e *GateEvent) Time() time.Time {
	return e.time
}

// Gate is a point in the world where new objects can appear
type Gate struct {
	Location   pixel.Vec
	Status     gateStatus
	Reserved   bool // gate is reserved by an object to be used next turn
	ReservedBy uuid.UUID

	// Wait this long before allowing a new spawn
	SpawnCoolDown time.Duration
	LastSpawn     time.Time

	Radius float64 // size

	eventNotifier behavior.EventNotifier

	Atlas *text.Atlas
}

// String returns the gate as string
func (g *Gate) String() string {
	return fmt.Sprintf("%#v", g.Location)
}

// CanSpawn returns true if the gate can spawn
func (g *Gate) CanSpawn() bool {
	switch {
	case g.Status != GateOpen:
		return false
	case g.Reserved:
		return false
	case time.Now().Sub(g.LastSpawn) < g.SpawnCoolDown:
		return false
	}
	return true
}

// Reserve reserves a gate if it's available
func (g *Gate) Reserve(id uuid.UUID) error {
	if g.CanSpawn() {
		g.Reserved = true
		g.ReservedBy = id
		g.LastSpawn = time.Now()
		g.eventNotifier.Notify(&GateEvent{
			description: fmt.Sprintf("gate [%v] reserved for [%v]", g, id),
			time:        time.Now(),
		})
		return nil
	}
	return fmt.Errorf("gate %#v already reserved or closed", g)
}

// Release removes a gates reservation
func (g *Gate) Release() {
	g.eventNotifier.Notify(&GateEvent{
		description: fmt.Sprintf("gate [%v] reservation released", g),
		time:        time.Now(),
	})
	g.Reserved = false
}

// Draw draws the gate on the screen
func (g *Gate) Draw(win *pixelgl.Window) {
	// TODO: Probably best to create ahead of time
	imd := imdraw.New(nil)

	if g.Reserved || g.Status == GateClosed {
		imd.Color = colornames.Red
	} else {
		imd.Color = colornames.Green
	}
	imd.Push(g.Location)
	imd.Circle(g.Radius, 2)
	imd.Draw(win)

	// remaining time until next spawn
	txt := text.New(g.Location, g.Atlas)
	txt.Color = colornames.Yellow

	label := ""

	switch {
	case g.Status == GateClosed:
		label = "inf"
	case !g.CanSpawn():
		label = fmt.Sprintf("%v", g.SpawnCoolDown-time.Now().Sub(g.LastSpawn).Truncate(time.Second))
	default:
		label = "inf"

	}

	fmt.Fprintf(txt, label)
	txt.Draw(win, pixel.IM)
}
