package world

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"gogs.wetsnow.com/dant/alphaville/observer"
	"gogs.wetsnow.com/dant/alphaville/utils"
)

// Stats keeps trackof world wide stats
// Implements observer.EventObserver interface
type Stats struct {
	Fps            int // frames per second
	ObjectsSpawned int // number of spawned objects
	Ups            int // updates (ticks) sper second
}

// NewStats returns a Stats object
func NewStats() *Stats {
	return &Stats{}
}

// String returns stats in a nice format
func (s *Stats) String() string {
	buf := bytes.NewBufferString("")

	fmt.Fprintf(buf, "\n=== World Stats ===\n")

	tmpl, err := template.New("stats").Parse(
		`  
  > Frames Per Second: {{.Fps}}
  > Updates Per Second: {{.Ups}}
  > Total Objects Spawned: {{.ObjectsSpawned}}
`)

	if err != nil {
		log.Fatalf("stats conversion error: %v", err)
	}
	err = tmpl.Execute(buf, s)
	if err != nil {
		log.Fatalf("stats conversion error: %v", err)
	}
	return buf.String()
}

func (s *Stats) processWorldEvent(e *worldEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "fps":
			s.Fps = utils.Atoi(data.Value)
		case "ups":
			s.Ups = utils.Atoi(data.Value)
		case "spawn":
			s.ObjectsSpawned++

		}
	}
}

func (s *Stats) processGateEvent(e *GateEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "none_yet":
			continue
		}
	}
}

// OnNotify runs when a notification is received
func (s *Stats) OnNotify(e observer.Event) {
	switch event := e.(type) {
	case nil:
		log.Printf("nil notification")
	case *worldEvent:
		s.processWorldEvent(event)
	case *GateEvent:
		s.processGateEvent(event)
	}
}
