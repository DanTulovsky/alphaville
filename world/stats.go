package world

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"

	"github.com/DanTulovsky/alphaville/observer"
	"github.com/DanTulovsky/alphaville/utils"
)

// Stats keeps track of world wide stats
// Implements observer.EventObserver interface
type Stats struct {
	Fps            int // frames per second
	ObjectsSpawned int // number of spawned objects
	Ups            int // updates (ticks) per second

	console io.ReadWriter
}

// NewStats returns a Stats object
func NewStats(console io.ReadWriter) *Stats {
	s := &Stats{}
	if console != nil {
		s.console = console
	} else {
		s.console = os.Stdout
	}
	return s
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
		}
	}
}

func (s *Stats) processGateEvent(e *GateEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "created":
			fmt.Fprintf(s.console, "gate [%v] created\n", data.Value)
		case "spawn":
			fmt.Fprintf(s.console, "gate spawned [%v]\n", data.Value)
			s.ObjectsSpawned++
		}
	}
}

func (s *Stats) processTargetEvent(e *TargetEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "created":
			// log.Printf("target [%v] created", data.Value)
		case "destroyed":
			// log.Printf("target [%v] destroyed", data.Value)
		}
	}
}

func (s *Stats) processObjectEvent(e *ObjectEvent) {
	for _, data := range e.Data() {
		switch data.Key {
		case "created":
			fmt.Fprintf(s.console, "object [%v] created\n", data.Value)
		case "target_found":
			fmt.Fprintf(s.console, "target [%v] found\n", data.Value)
		}
	}
}

// OnNotify runs when a notification is received
func (s *Stats) OnNotify(e observer.Event) {
	switch event := e.(type) {
	case nil:
		fmt.Fprintf(s.console, "nil notification\n")
	case *worldEvent:
		s.processWorldEvent(event)
	case *GateEvent:
		s.processGateEvent(event)
	case *TargetEvent:
		s.processTargetEvent(event)
	case *ObjectEvent:
		s.processObjectEvent(event)
	}
}

// Name returns the name of this object
func (s *Stats) Name() string {
	return "global_stats"
}
