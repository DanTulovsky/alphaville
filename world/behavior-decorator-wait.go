package world

import (
	"fmt"
	"time"

	"github.com/askft/go-behave/core"
)

// Delayer waits a specified amount of time and prepends ? to the name to signify the waiting.
func Delayer(params core.Params, child core.Node) core.Node {
	base := core.NewDecorator("Delayer", params, child)
	d := &delayer{Decorator: base}

	ms, err := params.GetInt("ms")
	if err != nil {
		panic(err)
	}

	d.delay = time.Duration(ms) * time.Millisecond
	return d
}

// delayer ...
type delayer struct {
	*core.Decorator
	delay     time.Duration // delay in milliseconds
	start     time.Time
	savedName string
}

// Enter ...
func (d *delayer) Enter(ctx *core.Context) {
	d.start = time.Now()
	o := ctx.Owner.(Object)
	d.savedName = o.Name()
	o.SetName(fmt.Sprintf("[?] %v", d.savedName))
}

// Tick ...
func (d *delayer) Tick(ctx *core.Context) core.Status {
	if time.Since(d.start) > d.delay {
		return core.Update(d.Child, ctx)
	}
	return core.StatusRunning
}

// Leave ...
func (d *delayer) Leave(ctx *core.Context) {

	o := ctx.Owner.(Object)
	o.SetName(d.savedName)
}
