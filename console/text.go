// Package console provides a text based console for debug and input
package console

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

// Layout sets the layout of the text console and gets run on every tick
func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if _, err := g.SetView("output", 0, 0, maxX-1, maxY-7); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	if _, err := g.SetView("input", 0, maxY-6, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
	}
	return nil
}

// Quit quits.
func Quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// CreateViews sets the layout of the text console
func CreateViews(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if output, err := g.SetView("output", 0, 0, maxX-1, maxY-7); err != nil {
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		output.Frame = true
		output.Autoscroll = true
		output.Title = "output"
	}
	if input, err := g.SetView("input", 0, maxY-6, maxX-1, maxY-1); err != nil {
		if err != nil && err != gocui.ErrUnknownView {
			return err
		}
		input.Editable = true
		input.Highlight = true
		input.Frame = true
		input.Title = "input"
		input.SetCursor(input.Origin())
	}
	return nil
}

func handleEnter(g *gocui.Gui, v *gocui.View) error {
	v.Rewind()

	ov, e := g.View("output")
	if e != nil {
		log.Println("Cannot get output view:", e)
		return e
	}

	_, e = fmt.Fprint(ov, v.Buffer())
	if e != nil {
		log.Println("Cannot print to output view:", e)
	}

	v.Clear()
	v.SetCursor(v.Origin())

	return nil
}

// New returns a new text gui
func New() *gocui.Gui {

	// text console for debug output and control
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}

	g.SetManagerFunc(Layout)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("input", gocui.KeyEnter, gocui.ModNone, handleEnter); err != nil {
		log.Panicln(err)
	}

	// g.Mouse = true
	g.Cursor = true

	return g
}
