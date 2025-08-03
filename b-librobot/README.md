# Robot Simulator Library

This library provides a simulation of a warehouse environment with robots that can be controlled to perform tasks.

## Features

*   Simulate multiple warehouses and robots.
*   Control robots with a simple command language.
*   Simulate crate handling in a warehouse environment.
*   Support for robots with diagonal movement.

## Installation

To install the library, use the following command:

```bash
go get github.com/ScottDWilson/robot-challenge-SDWilson/b-librobot/librobot
```

## Core Concepts

### Warehouse

A `Warehouse` represents the simulated warehouse environment. It provides a space where robots can operate. The `Warehouse` interface defines the following methods:

*   `Robots() []Robot`: Returns a list of all robots currently in the warehouse.

### Robot

A `Robot` represents a robot within the warehouse. Each robot can be given tasks to perform. The `Robot` interface defines the following methods:

*   `EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error)`: Adds a new task to the robot's queue. The `commands` string is a sequence of commands for the robot to execute. The method returns a `taskID`, a channel for position updates, and a channel for errors.
*   `CancelTask(taskID string) error`: Cancels a task by its `taskID`.
*   `CurrentState() RobotState`: Returns the current state of the robot.

Key features of a Robot:

*   Each robot has a unique ID.
*   Each robot operates within a specific warehouse.
*   Each robot maintains its own state, including its position and whether it is carrying a crate.
*   Robots execute tasks in a FIFO queue.

### RobotState

The `RobotState` struct represents the current state of a robot. It contains the following fields:

*   `X uint`: The X coordinate of the robot (0-GridSize).
*   `Y uint`: The Y coordinate of the robot (0-GridSize).
*   `HasCrate bool`: Whether the robot is currently carrying a crate.

### Tasks

Each robot operates by being given 'tasks' which each consist of a string of 'commands':

All of the commands to the robot consist of a single capital letter and different commands are optionally delineated by whitespace.

The robot should accept the following commands:

- N move one unit north
- W move one unit west
- E move one unit east
- S move one unit south

Example command sequences:

* The command sequence: `"N E S W"` will move the robot in a full square, returning it to where it started.

* If the robot starts in the south-west corner of the warehouse then the following commands will move it to the middle of the warehouse: `"N E N E N E N E"`

The robot will only perform a single task at a time: if additional tasks are given to the robot while is busy performing a task, those additional tasks are queued up, and will be executed once the preceding task is completed (or aborted for some reason).  Each task is identified with a unique string ID, and a task which is either in progress or enqueued can be aborted/cancelled at any time.  If the robot is unable to execute a particular command (for instance, because the command would cause the robot to run into the edges of the warehouse grid) then an error occurs, and the entire task is aborted.

## Diagonal Movement

To use diagonal movement, you must create a `DiagonalRobot` instead of a regular `Robot`.

```go
// Add a diagonal robot to warehouse
robot, err := librobot.AddDiagonalRobot(warehouse, 0, 0, "R1")
if err != nil {
    log.Printf("Unexpected error when adding robot; error: %v \n", err)
}
```

The robot will automatically combine pairs of orthogonal commands into diagonal movements. For example, "N E" will be combined into a single "↗" (NorthEast) command.

```go
command_string := "N E E N W W"
```

This will result in the following movements:

*   N E -> ↗ (North-East)
*   E
*   N W -> ↖ (North-West)
*   W

## Crate Handling

To use crate handling, you must create a `CrateWarehouse` instead of a regular `Warehouse`.

```go
// Create a crate warehouse
warehouse := librobot.NewCrateWarehouse()

// Add a crate to the warehouse
err := warehouse.AddCrate(1, 1)
if err != nil {
    // handle error
}
```

## Usage

Here's a basic example of how to use the library:

```go
package main

import (
	"fmt"
	"log"

	"github.com/ScottDWilson/robot-challenge-SDWilson/b-librobot/librobot"
)

func main() {
	// Create an empty warehouse
	warehouse1 := librobot.NewWarehouse()

	// Add 1 robot to warehouse
	robot1, err := librobot.AddRobot(warehouse1, 0, 0, "R1")
	if err != nil {
		log.Printf("Unexpected error when adding robot; error: %v \n", err)
	}

	// Send commands to robot queue using EqueueTask function
	command_string := "N N E E N E S" // Robot starting at 0,0 will end up at position 3,2
	taskID, _, _ := robot1.EnqueueTask(command_string)

	// Wait for completion...
	log.Println("Robot in action... waiting for completion...")

	// Check robot position
	robot1_state := robot1.CurrentState()
	if robot1_state.X != 3 || robot1_state.Y != 2 {
		log.Fatalf("Robot failed to move to position on task %v", taskID)
	}

	log.Println("Basic Usage Example completed successfully")
	fmt.Printf("Robot Position %v : %v", robot1_state.X, robot1_state.Y)
}
```

## Documentation

For more detailed documentation, please refer to the Go docs:

```bash
go doc github.com/ScottDWilson/robot-challenge-SDWilson/b-librobot/librobot
```

You can also view the documentation online at [GoDoc](https://pkg.go.dev/github.com/ScottDWilson/robot-challenge-SDWilson/b-librobot/librobot). (Not actually available for interview)

## Running Tests

To run the tests, use the following command:

```bash
go test github.com/ScottDWilson/robot-challenge-SDWilson/b-librobot/librobot
```

### Test Cover

To run and view test cover; run 

```bash
cd /b-librobot/librobot
go test -coverprofile='coverage.out' ./...
go tool cover -html='coverage.out'
```

## Errors

The library defines several errors that can be returned by the functions. These errors are defined in the `librobot_errors.go` file.

*   `ErrOutOfBounds`: Returned when the robot attempts to move out of bounds.
*   `ErrRobotHasCrate`: Returned when the robot attempts to grab a crate while already carrying one.
*   `ErrCrateNotFound`: Returned when the robot attempts to grab a crate that does not exist.
*   `ErrCrateExists`: Returned when attempting to add a crate to a location where a crate already exists.
*   `ErrRobotNotCrate`: Returned when the robot attempts to drop a crate when it is not carrying one.
*   `ErrInvalidWarehouseType`: Returned when attempting to perform an operation on the wrong type of warehouse.

## Contributing

Contributions are welcome! Please submit a pull request with your changes.


# Original Instructions:

We wish to create a simulator which mimics the behaviour of our new robots.

The simulation should take the form of a Golang library with associated documentation and tests.

## Library Interface

The simulator should implement the following interfaces (see the included file):

```
type Warehouse interface {
	Robots() []Robot
}

type Robot interface {
	EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) 

	CancelTask(taskID string) error

	CurrentState() RobotState
}

type RobotState struct {
	X uint
	Y uint
	HasCrate bool
}
```

## Requirements

### Part One

Implement the simulator so that you can create an instance of a simulated `Warehouse`, then add one or more `Robots` to it, and issue instructions to those Robots.

Some notes:
* Only one robot should be able to occupy a location within the warehouse at a time.
* Multiple robots may operate within a single warehouse.
* Multiple warehouses may be simulated at a time.
* Each robot should take one second of real time to perform each command.

Provide documentation and tests to allow users of library to use the simulator and validate its correct operation.

### Part Two

Now, we add a lifting claw to the robot, so that it can move crates in the warehouse.

Add a new interface:

```
type CrateWarehouse interface {
	Warehouse

	AddCrate(x uint, y uint) error
	DelCrate(x uint, y uint) error
}
```

Then extend the valid commands supported by the robot simulator to include the following:
* "G" - If the robot is at a location with a crate, grab it.
* "D" - Drop a carried crate at the robot's current position.

Some notes:
* The robot should only be able to carry one crate at a time.
* A crate may not be dropped at a location where there is already a crate.

Provide tests to validate the correct simulation of crate handling.

### Part Three

Now, we wish to extend the simulation to allow representation a new kind of robot which is able to travel diagnally when traversing the warehouse grid:

The supported command syntax for the simulated robot should remain the same, but if the robot is issued a pair of commands which would result in it moving (for example) North and then East, it should instead simply perform a single North-East movement.

Provide tests to validate that the new simulated robot performs correctly.
