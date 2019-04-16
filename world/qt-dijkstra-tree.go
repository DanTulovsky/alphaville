//  dijkstra is an highly optimized implementation of the Dijkstra
// From: https://github.com/albertorestifo/dijkstra

package world

import (
	"fmt"

	"github.com/faiface/pixel"
)

type node struct {
	key  *Node
	cost int
}

// PathFinder is a path finder algorithm
type PathFinder interface {
	Path(t *Tree, start, target pixel.Vec) (path NodeList, cost int, err error)
}

// DijkstraPathFinder implements Dijkstra path finding
type DijkstraPathFinder struct {
}

// Graph is a rappresentation of how the points in our graph are connected
// between each other
// type Graph map[string]map[string]int

// Path finds the shortest path between start and target, also returning the
// total cost of the found path.
func (d *DijkstraPathFinder) Path(t *Tree, start, target pixel.Vec) (path NodeList, cost int, err error) {
	if len(t.Leaves) == 0 {
		err = fmt.Errorf("cannot find path in empty graph")
		return
	}

	// ensure start and target are part of the graph
	startNode, err := t.Locate(start)
	if err != nil {
		err = fmt.Errorf("cannot find start %v in graph: %v", start, err)
		return
	}
	if _, er := t.Locate(target); er != nil {
		err = fmt.Errorf("cannot find target %v in graph: %v", target, er)
		return
	}

	explored := make(map[*Node]bool)  // set of nodes we already explored
	frontier := NewQueue()            // queue of the nodes to explore
	previous := make(map[*Node]*Node) // previously visited node

	// add starting point to the frontier as it'll be the first node visited
	frontier.Set(startNode, 0)

	// run until we visited every node in the frontier
	for !frontier.IsEmpty() {
		// get the node in the frontier with the lowest cost (or priority)
		aKey, aPriority := frontier.Next()
		n := node{aKey, aPriority}
		// fmt.Printf("%#+v\n", n.key)

		// when the node with the lowest cost in the frontier is target, we can
		// compute the cost and path and exit the loop
		if n.key.bounds.Center() == target {
			cost = n.cost

			nKey := n.key
			for nKey.bounds.Center() != start {
				path = append(path, nKey)
				nKey = previous[nKey]
			}
			break
		}

		// add the current node to the explored set
		explored[n.key] = true

		// loop all the neighboring nodes
		// for _, nKey := range g.edges[*n.key] {
		for _, nKey := range n.key.Neighbors() {
			// skip already-explored nodes
			if explored[nKey] {
				continue
			}
			// cost to get to this node is the length of the line
			nCost := int(nKey.bounds.Center().Sub(n.key.bounds.Center()).Len())
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

	// Do not add the origin, as this makes sometimes move backwards
	// and leads to meeting ones getting stuck for long periods of time

	// add the origin at the end of the path
	// path = append(path, startNode)
	// path = append(path, NewItemNode(uuid.New(), start, 0))

	// reverse the path because it was popilated
	// in reverse, form target to start
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	// path = append(path, NewItemNode(uuid.New(), target, 0))
	return
}
