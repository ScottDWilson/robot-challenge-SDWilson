package librobot

import "errors"

// Errors for test validation
var (
	// ErrOutOfBounds indicates that a command would move the robot outside the warehouse grid.
	ErrOutOfBounds = errors.New("command would move robot out of bounds")
	// ErrPositionOccupied indicates that the target position is already occupied by another robot.
	ErrPositionOccupied = errors.New("target position already occupied by another robot")
	// ErrRobotNotFound indicates that a specified robot ID was not found in the warehouse.
	ErrRobotNotFound = errors.New("robot not found") // Currently not returned by any public method, but useful if we add GetRobot(id string)
	// ErrTaskNotFound indicates that a specified task ID was not found for the robot.
	ErrTaskNotFound = errors.New("task not found")
	// ErrCrateNotFound indicates that no crate exists at the specified location.
	ErrCrateNotFound = errors.New("crate not found at specified location")
	// ErrCrateExists indicates that a crate already exists at the specified location.
	ErrCrateExists = errors.New("crate already exists at specified location")
	// ErrInvalidWarehouseType indicates that an operation was attempted on an incompatible warehouse type.
	ErrInvalidWarehouseType = errors.New("invalid warehouse type")
	// ErrRobotHasCrate indicates that the robot already carries a crate
	ErrRobotHasCrate = errors.New("robot is already carrying a crate")
	// ErrRobotNotCrate indicates that the robot already carries a crate
	ErrRobotNotCrate = errors.New("robot is not carrying a crate")
)
