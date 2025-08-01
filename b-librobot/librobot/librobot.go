package librobot

import "time"

// Constants used for simulation
const (
	// GridSize defines the dimension of the square warehouse grid (for example; 10x10).
	// Coordinates range from 0 to GridSize.
	GridSize = 10
	// CommandExecutionTime defines the real time taken to execute one command.
	CommandExecutionTime = 1 * time.Second
)

// Warehouse provides an abstraction of a simulated warehouse containing robots.
type Warehouse interface {
	Robots() []Robot
}

// CrateWarehouse provides an abstraction of a simulated warehouse containing both robots and crates.
type CrateWarehouse interface {
	Warehouse

	AddCrate(x uint, y uint) error
	DelCrate(x uint, y uint) error
}

// Robot provides an abstraction of a warehouse robot which accepts tasks in the form of strings of commands.
type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error)

	CancelTask(taskID string) error

	CurrentState() RobotState
}

// RobotState provides an abstraction of the state of a warehouse robot.
type RobotState struct {
	X        uint // X coordinate of the robot (0-GridSize)
	Y        uint // Y coordinate of the robot (0-GridSize)
	HasCrate bool // Whether the robot is currently carrying a crate
}

// ALL DONE.
