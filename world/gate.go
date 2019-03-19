package world

import (
	"fmt"
	"time"

	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"github.com/google/uuid"
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

	Atlas *text.Atlas
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
		return nil
	}
	return fmt.Errorf("gate %#v already reserved or closed", g)
}

// UnReserve removes a gates reservation
func (g *Gate) UnReserve() {
	g.Reserved = false
}

// Draw draws the gate on the screen
func (g *Gate) Draw(win *pixelgl.Window) {
	// TODO: Probably best to create ahead of time
	imd := imdraw.New(nil)

	// TODO: draw open and closed gates differently
	if g.Reserved {
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
	case !g.CanSpawn():
		label = fmt.Sprintf("%v", g.SpawnCoolDown-time.Now().Sub(g.LastSpawn).Truncate(time.Second))
	default:
		label = "∞"

	}

	fmt.Fprintf(txt, label)
	txt.Draw(win, pixel.IM)
}
