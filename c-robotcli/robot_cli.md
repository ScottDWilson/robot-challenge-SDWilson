# Robot Simulator CLI User Guide

This is an interactive command-line application for simulating a robot warehouse, now built with the Cobra framework. The application provides a structured command-line interface for managing robots and their tasks, with a real-time visualization of the simulation state.

## Quick Start

Create a project directory, for example robot-cli.

Inside this directory, create a subdirectory librobot and place the librobot/librobot.go file there.

Place the main.go file in the main robot-sim directory.

Initialize the Go module and get the Cobra dependency:

go mod init robot-cli
go get github.com/spf13/cobra

Compile and run the application from your terminal:

go run .

The application will start, and you can now use the commands below.

## Available Commands
### `add_robot`

Adds a new robot to the warehouse.

**Usage:**

```bash
robot-cli add_robot <id> <x> <y>
```

-   `<id>`: A unique identifier for the robot (e.g., `R1`).
-   `<x>`: The initial X coordinate (0-9).
-   `<y>`: The initial Y coordinate (0-9).

The id of the robot will be used as the label on view

**Example:**

```bash
robot-cli add_robot R0 5 5
```

### `add_diag_robot`

Adds a new diagonal robot to the warehouse. Diagonal robots can move diagonally in addition to the cardinal directions. Diagonal robots respond to the normal list of commands and will automatically optimise the route by combining tasks if suitable (For example; N E will become a single North-East command)

**Usage:**

```bash
robot-cli add_diag_robot <id> <x> <y>
```

-   `<id>`: A unique identifier for the robot (e.g., `R1`).
-   `<x>`: The initial X coordinate (0-9).
-   `<y>`: The initial Y coordinate (0-9).

**Example:**

```bash
robot-cli add_diag_robot D2D2 2 2
```

### `add_task`

Enqueues a task for a robot. The task is a string of single-character commands that the robot executes sequentially.

**Usage:**

```bash
robot-cli add_task <robot_id> <commands>
```

-   `<robot_id>`: The ID of the robot.
-   `<commands>`: A string of commands. Available commands are:
    -   `N`: Move North (up)
    -   `E`: Move East (right)
    -   `S`: Move South (down)
    -   `W`: Move West (left)
    -   `G`: Pickup a crate at the current location. Only picks a crate if it exists.
    -   `D`: Drop a crate at the current location. Only drops a crate if one does not already exist.

**Example:**

```bash
robot-cli add_task R2 NNNWWWGND
```

### `add_crate`

Adds a stationary crate to the warehouse at a specific location.

**Usage:**

```bash
robot-cli add_crate <x> <y>
```

-   `<x>`: The X coordinate (0-9).
-   `<y>`: The Y coordinate (0-9).

**Example:**

```bash
robot-cli add_crate 5 7
```

### `del_crate`

Deletes a crate from the warehouse at a specific location.

**Usage:**

```bash
robot-cli del_crate <x> <y>
```

-   `<x>`: The X coordinate (0-9).
-   `<y>`: The Y coordinate (0-9).

**Example:**

```bash
robot-cli del_crate 5 7
```

### `cancel_task`

Cancels a running task.

**Usage:**

```bash
robot-cli cancel_task <robot_id> <task_id>
```

-   `<robot_id>`: The ID of the robot with the task.
-   `<task_id>`: The unique ID returned when the task was enqueued.

**Example:**

```bash
robot-cli cancel_task R2 1678881234567890
```

### `view`

Displays a real-time ASCII view of the warehouse.

**Usage:**

```bash
robot-cli view
```

This command starts a visualization in a separate goroutine, updating the terminal with the current state of the warehouse, including robot positions and crate locations.

Locations are marked as follows:
-   `[C]`: A crate is at this location.
-   `R~`: A robot is at this location, for example 'R0'
-   `R-*`: A robot is carrying a crate at this location, for example 'R0*'
-   `R-_`: A robot and a crate is at this location, for example 'R0_'

### `stop_view`

Stops the real-time ASCII view.

**Usage:**

```bash
robot-cli stop_view
```

### `exit`

Stops the cli and closes the program

**Usage:**

```bash
robot-cli exit
```

This command halts the rendering of the warehouse state.

## Simulation Visualization

The application provides a real-time, text-based grid to show the state of the warehouse.

-   **Grid:** The warehouse is a 10x10 grid.
-   **Robots:** Robots are represented by the letter `R`.
-   **Crates:** Crates are represented by the letter `C`.

Locations are marked as follows:
-   `[C]`: A crate is at this location.
-   `R~`: A robot is at this location, for example 'R0'
-   `R-*`: A robot is carrying a crate at this location, for example 'R0*'
-   `R-_`: A robot and a crate is at this location, for example 'R0_'

## Interactive Mode

If no command-line arguments are provided, the CLI starts in interactive mode. In this mode, you can enter commands at the prompt, and the simulation will update accordingly.

To exit the interactive mode, type `exit`.
