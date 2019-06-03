package world

import (
	"bytes"
	"html/template"
	"log"

	behave "github.com/askft/go-behave"
	"github.com/askft/go-behave/core"
	"github.com/askft/go-behave/store"

	. "github.com/askft/go-behave/common/decorator"

	"github.com/faiface/pixel/pixelgl"
)

// WonderBehavior randomly wonders around the world. Uses Behavior Trees.
type WondererBehavior struct {
	description string
	name        string
	parent      Object
	t           *behave.BehaviorTree
}

// NewWondererBehavior return a WondererBehavior
func NewWondererBehavior(f PathFinder, parent Object) *WondererBehavior {
	b := &WondererBehavior{
		name:        "wonderer_behavior",
		description: "wonders aimlessly...",
		parent:      parent,
	}

	// behavior tree itself
	root := Repeater(core.Params{"n": 0}, nil)

	cfg := behave.Config{
		Owner: b.parent,
		Data:  store.NewBlackboard(),
		Root:  root,
	}

	t, err := behave.NewBehaviorTree(cfg)
	if err != nil {
		log.Fatalf("error creating behavior tree: %v", err)
	}

	b.t = t
	return b
}

// String returns ...
func (b *WondererBehavior) String() string {
	buf := bytes.NewBufferString("")
	tmpl, err := template.New("physObject").Parse(
		`
Behavior
  Name: {{.Name}}	
  Desc: {{.Description}}	
`)

	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}
	err = tmpl.Execute(buf, b)
	if err != nil {
		log.Fatalf("behavior conversion error: %v", err)
	}

	return buf.String()
}

// Name returns the name of the behavior
func (b *WondererBehavior) Name() string {
	return b.name
}

// Parent returns the parent object of the behavior
func (b *WondererBehavior) Parent() Object {
	return b.parent
}

// SetParent returns the parent object of the behavior
func (b *WondererBehavior) SetParent(p Object) {
	b.parent = p
}

// Description returns the name of the behavior
func (b *WondererBehavior) Description() string {
	return b.description
}

func (b *WondererBehavior) Update(w *World, o Object) {
}

// Draw draws any artifacts of the behavior
func (b *WondererBehavior) Draw(win *pixelgl.Window) {

}
