package world

import (
	"log"

	"gogs.wetsnow.com/dant/alphaville/behavior"
)

// Stats keeps trackof world wide stats
// Implements behavior.EventObserver interface
type Stats struct {
}

// NewStats returns a Stats object
func NewStats() *Stats {
	return &Stats{}
}

// OnNotify runs when a notification is received
func (s *Stats) OnNotify(e behavior.Event) {
	log.Printf("NOTIFICATION: %v", e)
}
