package main

import (
	"log"

	"github.com/faiface/pixel"

	"gogs.wetsnow.com/dant/alphaville/graph"
)

func main() {
	g := graph.NewGraph()
	log.Printf("%v", g)

	n1 := graph.NewItemNode(pixel.V(0, 0))
	n2 := graph.NewItemNode(pixel.V(10, 5))
	n3 := graph.NewItemNode(pixel.V(45, 2))
	n4 := graph.NewItemNode(pixel.V(23, 7))
	n5 := graph.NewItemNode(pixel.V(3, 19))
	n6 := graph.NewItemNode(pixel.V(100, 20))

	g.AddNode(n1)
	g.AddNode(n2)
	g.AddNode(n3)
	g.AddNode(n4)
	g.AddNode(n5)
	g.AddNode(n6)
	log.Printf("%v", g)

	g.AddEdge(n1, n2)
	g.AddEdge(n3, n4)
	g.AddEdge(n5, n6)
	g.AddEdge(n1, n4)
	g.AddEdge(n1, n3)
	g.AddEdge(n3, n5)
	log.Printf("%v", g)
}
