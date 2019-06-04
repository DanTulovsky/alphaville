package world

import (
	"bytes"
	"html/template"
	"log"

	behave "github.com/askft/go-behave"
	"github.com/askft/go-behave/core"

	action "github.com/askft/go-behave/common/action"
	composite "github.com/askft/go-behave/common/composite"
	decorator "github.com/askft/go-behave/common/decorator"
)

// WonderBehavior randomly wonders around the world. Uses Behavior Trees.
type WondererBehavior struct {
	DefaultBehavior
}

// NewWondererBehavior return a WondererBehavior
func NewWondererBehavior(parent Object, w *World) *WondererBehavior {
	// behavior tree itself
	root := decorator.Repeater(core.Params{"n": 0},
		composite.Sequence(
			Delayer(core.Params{"ms": 3000}, // think about what to do
				action.Succeed(nil, nil)),
			Wonder(core.Params{}, nil),
		))

	b := &WondererBehavior{
		DefaultBehavior: DefaultBehavior{
			name:        "wonderer_behavior",
			description: "wonders aimlessly...",
			parent:      parent,
		},
	}

	cfg := behave.Config{
		Owner: b.parent,
		Data:  w, // for now the world is the Data
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

func (b *WondererBehavior) Update(w *World, o Object) {
	b.t.Update()
	// util.PrintTreeInColor(b.t.Root)
}
