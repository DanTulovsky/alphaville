package world

// Type defines a type for everything in the world
// Anything appearing in the world should have a type
type Type int

const (
	unknownType = iota

	// the world itself
	worldType

	// gates allow objects to enter the world
	gateType

	// rectangular objects
	objectRectType

	// circular objects
	objectCircleType

	// ellipse objects
	objectEllipseType

	// ground
	groundType

	// fixture (wall, etc...)
	fixtureType
)
