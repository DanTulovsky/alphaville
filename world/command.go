package world

import "log"

// Command is a command executed by the object
type Command interface {
	Execute()
}

// StopCommand tells the object to stop moving
type StopCommand struct {
}

// Execute runs the command
func (c *StopCommand) Execute(o Object) {
	log.Printf("stop %v", o.Name())
}

// MoveCommand tells the object to move
type MoveCommand struct {
}

// Execute runs the command
func (c *MoveCommand) Execute(o Object) {
	log.Printf("move %v", o.Name())
}

// NewCommand returns a new command
// func NewCommand(o Object, m func()) *Command {
// 	return &Command{object: o, method: m}
// }
