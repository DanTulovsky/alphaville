package world

import (
	"fmt"

	"github.com/faiface/pixel/pixelgl"
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
}

// Reserve reserves a gate if it's available
func (g *Gate) Reserve(id uuid.UUID) error {
	if !g.Reserved && g.Status == GateOpen {
		g.Reserved = true
		g.ReservedBy = id
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
	imd.Circle(10, 2)
	imd.Draw(win)
}
