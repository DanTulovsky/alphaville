package main

import (
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/quadtree"
	"golang.org/x/image/colornames"
)

func qt() {

	start := pixel.R(11, 11, 12, 12)
	target := pixel.R(790, 790, 791, 791)

	objects := []pixel.Rect{
		start,
		target,
		// pixel.R(0, 0, 10, 10),
		// pixel.R(20, 20, 30, 30),
		pixel.R(20, 400, 40, 600),
		pixel.R(10, 40, 50, 50),
		pixel.R(100, 40, 150, 50),
		pixel.R(0, 400, 500, 405),
		pixel.R(511, 400, 800, 405),
		pixel.R(111, 200, 800, 205),
		pixel.R(650, 405, 670, 760),
		pixel.R(650, 525, 670, 525),
		pixel.R(100, 100, 150, 280),
		pixel.R(500, 500, 650, 780),
		pixel.R(100, 100, 200, 200),
		pixel.R(150, 150, 250, 250),
	}

	var width float64 = 800
	var height float64 = 860

	bounds := pixel.R(0, 0, width, height)

	// minimum size of the bounding square in the quadtree
	var minSize float64 = 10

	qt, err := quadtree.NewTree(bounds, objects, minSize)
	if err != nil {
		log.Fatalf("%v", err)
	}

	g := qt.ToGraph(start, target)

	cfg := pixelgl.WindowConfig{
		Title:  "Quadtree",
		Bounds: pixel.R(0, 0, width, height),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	startNode := qt.Locate(start)
	targetNode := qt.Locate(target)
	path, _, err := graph.DijkstraPath(g, startNode.Bounds().Center(), targetNode.Bounds().Center())
	if err != nil {
		log.Printf("%v", err)
	}

	// Main loop to keep window running
	for !win.Closed() {
		// render below here
		win.Clear(colornames.Lightgreen)
		draw(win, qt, g, path)
		win.Update()
	}
}

func draw(win *pixelgl.Window, qt *quadtree.Tree, g *graph.Graph, path []*graph.Node) {
	drawTree, drawText, drawObjects := true, false, true
	qt.Draw(win, drawTree, drawText, drawObjects)
	graph.DrawPath(win, path)
}
