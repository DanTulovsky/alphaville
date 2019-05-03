package world

import (
	"strconv"
	"strings"

	"github.com/jroimartin/gocui"
	"github.com/tevino/abool"
)

// QuadTreeDebug contains variables for quadtree debuggin
type QuadTreeDebug struct {
	DrawTree    *abool.AtomicBool // draws the grid of the graph generated from the tree
	ColorTree   *abool.AtomicBool // colors the quadrants (white or black)
	DrawText    *abool.AtomicBool // draws the coordinates of the quadrants
	DrawObjects *abool.AtomicBool // draws outline of objects
}

// DebugConfig contains variables to turn on debugging
type DebugConfig struct {
	QT QuadTreeDebug
}

func (w *World) processDebugQTCommand(tokens []string, out *gocui.View) {
	// debug qt variable value
	v := strings.TrimSpace(tokens[0])
	b, _ := strconv.ParseBool(strings.TrimSpace(tokens[1]))

	switch v {
	case "draw_tree":
		w.debug.QT.DrawTree.SetTo(b)
	case "color_tree":
		w.debug.QT.ColorTree.SetTo(b)
	case "draw_text":
		w.debug.QT.DrawText.SetTo(b)
	case "draw_objects":
		w.debug.QT.DrawObjects.SetTo(b)
	}

}

func (w *World) processDebugCommand(tokens []string, out *gocui.View) {
	// debug world|qt variable value

	switch tokens[0] {
	case "world":
		// nothing yet
	case "qt":
		if len(tokens) == 3 {
			w.processDebugQTCommand(tokens[1:], out)
		}
	}

}
