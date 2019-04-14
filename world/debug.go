package world

// QuadTreeDebug contains variables for quadtree debuggin
type QuadTreeDebug struct {
	DrawTree    bool // draws the grid of the graph generated from the tree
	ColorTree   bool // colors the quadrants (white or black)
	DrawText    bool // draws the coordinates of the quadrants
	DrawObjects bool // draws outline of objects
}

// DebugConfig contains variables to turn on debugging
type DebugConfig struct {
	QT QuadTreeDebug
}
