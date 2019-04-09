/*
From: https://github.com/albertorestifo/dijkstra

Package dijkstra is an highly optimised implementation of the Dijkstra
algorithm, used for find the shortest path between points of a graph.

A graph is a map of points and map to the neighbouring points in the graph and
the cost to reach them.
A trivial example of a graph definition is:

	Graph{
		"a": {"b": 10, "c": 20},
		"b": {"a": 50},
		"c": {"b": 10, "a": 25},
	}

*/
package graph

import (
	"fmt"

	"github.com/faiface/pixel"
	"github.com/google/uuid"
)

type node struct {
	key  *Node
	cost int
}

// Graph is a rappresentation of how the points in our graph are connected
// between each other
// type Graph map[string]map[string]int

// DijkstraPath finds the shortest path between start and target, also returning the
// total cost of the found path.
func DijkstraPath(g *Graph, start, target pixel.Vec) (path []*Node, cost int, err error) {
	if len(g.nodes) == 0 {
		err = fmt.Errorf("cannot find path in empty graph")
		return
	}

	// ensure start and target are part of the graph
	if g.FindNode(start) == nil {
		err = fmt.Errorf("cannot find start %v in graph", start)
		return
	}
	if g.FindNode(target) == nil {
		err = fmt.Errorf("cannot find target %v in graph", target)
		return
	}

	explored := make(map[*Node]bool)  // set of nodes we already explored
	frontier := NewQueue()            // queue of the nodes to explore
	previous := make(map[*Node]*Node) // previously visited node

	// add starting point to the frontier as it'll be the first node visited
	frontier.Set(g.FindNode(start), 0)

	// run until we visited every node in the frontier
	for !frontier.IsEmpty() {
		// get the node in the frontier with the lowest cost (or priority)
		aKey, aPriority := frontier.Next()
		n := node{aKey, aPriority}
		// fmt.Printf("%#+v\n", n.key)

		// when the node with the lowest cost in the frontier is target, we can
		// compute the cost and path and exit the loop
		if n.key.value.V == target {
			cost = n.cost

			nKey := n.key
			for nKey.value.V != start {
				path = append(path, nKey)
				nKey = previous[nKey]
			}
			break
		}

		// add the current node to the explored set
		explored[n.key] = true

		// loop all the neighboring nodes
		for _, nKey := range g.edges[*n.key] {
			// skip already-explored nodes
			if explored[nKey] {
				continue
			}
			// cost to get to this node is the length of the line
			nCost := int(nKey.Value().V.Sub(n.key.Value().V).Len())
			// nCost := nKey.cost

			// if the node is not yet in the frontier add it with the cost
			if _, ok := frontier.Get(nKey); !ok {
				previous[nKey] = n.key
				frontier.Set(nKey, n.cost+nCost)
				continue
			}

			frontierCost, _ := frontier.Get(nKey)
			nodeCost := n.cost + nCost

			// only update the cost of this node in the frontier when
			// it's below what's currently set
			if nodeCost < frontierCost {
				previous[nKey] = n.key
				frontier.Set(nKey, nodeCost)
			}
		}
	}

	if len(path) == 0 {
		err = fmt.Errorf("Unable to find path from %v to %v", start, target)
		return
	}

	// add the origin at the end of the path
	path = append(path, g.FindNode(start))
	path = append(path, NewItemNode(uuid.New(), start, 0))

	// reverse the path because it was popilated
	// in reverse, form target to start
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	path = append(path, NewItemNode(uuid.New(), target, 0))
	return
}
