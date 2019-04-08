package main

import (
	"fmt"
	"log"

	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/google/uuid"
	"gogs.wetsnow.com/dant/alphaville/graph"
	"gogs.wetsnow.com/dant/alphaville/quadtree"
	"golang.org/x/image/colornames"
)

func run() {

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	start := pixel.R(11, 11, 12, 12)
	target := pixel.R(950, 950, 951, 951)

	objects := []pixel.Rect{
		start,
		target,
		pixel.R(0, 0, 10, 10),
		pixel.R(20, 20, 30, 30),
		pixel.R(10, 40, 50, 50),
		pixel.R(100, 40, 150, 50),
		pixel.R(400, 400, 405, 405),
		pixel.R(500, 500, 650, 780),
		pixel.R(700, 850, 850, 880),
	}

	bounds := pixel.R(0, 0, 1000, 1000)

	// minimum size of the bounding square in the quadtree
	var minSize float64 = 6

	qt, err := quadtree.NewTree(bounds, objects, minSize)
	if err != nil {
		log.Fatalf("%v", err)
	}

	nodeNeighbors := make(map[*quadtree.Node]quadtree.NodeList)

	log.Printf("%v", qt)

	perNode := func(n *quadtree.Node) {
		neighbors := n.Neighbors()
		nodeNeighbors[n] = neighbors

		fmt.Printf("Node: %v", n.Bounds())
		fmt.Println("  neighbors: ")
		for _, neighbor := range neighbors {
			fmt.Printf("    %v (%v)", neighbor.Bounds(), neighbor.Color())
			fmt.Println("")

		}
	}

	// print out the tree
	qt.ForEachLeaf(quadtree.Gray, perNode)

	g := graph.New()
	for node := range nodeNeighbors {
		gnode := graph.NewItemNode(uuid.New(), node.Bounds().Center(), 1)
		g.AddNode(gnode)
	}

	for node, neighbors := range nodeNeighbors {
		gnode := g.FindNode(node.Bounds().Center())
		for _, n := range neighbors {
			gneighbor := g.FindNode(n.Bounds().Center())
			g.AddEdge(gnode, gneighbor)
		}
		g.AddNode(gnode)
	}

	log.Printf("%v", g)

	cfg := pixelgl.WindowConfig{
		Title:  "Pixel Rocks!",
		Bounds: pixel.R(0, 0, 1000, 1000),
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	startNode := qt.Locate(start)
	targetNode := qt.Locate(target)
	log.Printf("start: %v\n", start.Center())
	log.Printf("  node: %v\n", startNode.Bounds())
	log.Printf("target: %v\n", target.Center())
	log.Printf("  node: %v\n", targetNode.Bounds())

	// add edges based on visibility from start and target to other nodes

	// Main loop to keep window running
	for !win.Closed() {
		// render below here
		win.Clear(colornames.Black)
		draw(win, qt, g, startNode, targetNode)
		win.Update()
	}
}

func draw(win *pixelgl.Window, qt *quadtree.Tree, g *graph.Graph, start, target *quadtree.Node) {
	imd := imdraw.New(nil)
	imd.Color = colornames.Red

	// draw the quadtree itself
	rectangles := []pixel.Rect{}

	perNode := func(n *quadtree.Node) {
		rectangles = append(rectangles, n.Bounds())
	}
	qt.ForEachLeaf(quadtree.Gray, perNode)

	for _, r := range rectangles {
		imd.Push(r.Min)
		imd.Push(r.Max)
		imd.Rectangle(1)
	}

	// draw the objects
	imd.Color = colornames.Yellow

	for _, r := range qt.Root().Objects() {
		imd.Push(r.Min)
		imd.Push(r.Max)
		imd.Rectangle(1)
	}

	// draw the path
	path, _, err := graph.DijkstraPath(g, start.Bounds().Center(), target.Bounds().Center())
	if err != nil {
		log.Fatalf("%v", err)
	}

	// log.Printf(">>> [path] %v", path)

	imd.Color = colornames.Lightblue
	for _, p := range path {
		imd.Push(p.Value().V)
	}
	imd.Line(1)

	imd.Draw(win)
}

func main() {
	pixelgl.Run(run)
}
