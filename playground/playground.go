package main

import (
	"fmt"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/quadtree"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
)

func run() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	start := pixel.R(11, 11, 12, 12)
	target := pixel.R(790, 790, 791, 791)

	objects := []pixel.Rect{
		start,
		target,
		// pixel.R(0, 0, 10, 10),
		// pixel.R(20, 20, 30, 30),
		// pixel.R(20, 400, 40, 600),
		// pixel.R(10, 40, 50, 50),
		// pixel.R(100, 40, 150, 50),
		// pixel.R(0, 400, 500, 405),
		// pixel.R(511, 400, 800, 405),
		pixel.R(111, 200, 800, 205),
		// pixel.R(650, 405, 670, 760),
		// pixel.R(650, 525, 670, 525),
		// pixel.R(100, 100, 150, 280),
		// pixel.R(500, 500, 650, 780),
		// pixel.R(100, 100, 200, 200),
		// pixel.R(150, 150, 250, 250),
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
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, width, height),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	// add edges based on visibility from start and target to other nodes
	atlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)

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
		draw(win, qt, g, atlas, path)
		win.Update()
	}
}

func draw(win *pixelgl.Window, qt *quadtree.Tree, g *graph.Graph, atlas *text.Atlas, path []*graph.Node) {
	imd := imdraw.New(nil)

	// draw the quadtree itself
	rectangles := quadtree.NodeList{}

	perNode := func(n *quadtree.Node) {
		rectangles = append(rectangles, n)
	}
	qt.ForEachLeaf(colornames.Gray, perNode)

	for _, r := range rectangles {
		imd = imdraw.New(nil)
		imd.Color = r.Color()
		imd.Push(r.Bounds().Min)
		imd.Push(r.Bounds().Max)
		imd.Rectangle(0)

		imd.Color = colornames.Red
		imd.Push(r.Bounds().Min)
		imd.Push(r.Bounds().Max)
		imd.Rectangle(1)
		imd.Draw(win)

		txt := text.New(r.Bounds().Center(), atlas)
		txt.Color = colornames.Darkgray
		label := fmt.Sprintf("%v,\n%v", r.Bounds().Center().X, r.Bounds().Center().Y)
		txt.Dot.X -= txt.BoundsOf(label).W() / 2
		fmt.Fprintf(txt, "%v", label)
		txt.Draw(win, pixel.IM)
	}

	imd = imdraw.New(nil)
	// draw the objects
	imd.Color = colornames.Yellow

	for _, r := range qt.Root().Objects() {
		imd.Push(r.Min)
		imd.Push(r.Max)
		imd.Rectangle(2)
	}

	imd.Color = colornames.Darkblue
	for _, p := range path {
		imd.Push(p.Value().V)
	}
	imd.Line(1)

	imd.Draw(win)
}

func main() {
	pixelgl.Run(run)
}
