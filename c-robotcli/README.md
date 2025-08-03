# Robot Simulator CLI

## Overview

The Robot Simulator CLI is an interactive command-line application for simulating a robot warehouse. Built using the Cobra framework, it provides a structured interface for managing robots, assigning tasks, and visualizing the simulation in real-time. This tool is designed for testing and demonstration purposes, allowing users to interact with the `librobot` simulator.

## Installation

1.  **Prerequisites:** Ensure you have Go installed on your system. You can download it from [go.dev](https://go.dev/dl/).

2.  **Clone the repository:**

    ```bash
    git clone github.com/ScottDWilson/robot-challenge-SDWilson
    cd robot-challenge-SDWilson/c-robotcli
    ```

3.  **Initialize the Go module:**

    ```bash
    go mod init robot-cli
    ```

4.  **Get the Cobra dependency:**

    ```bash
    go get github.com/spf13/cobra
    ```

## Usage

To run the CLI, navigate to the `c-robotcli` directory and execute:

```bash
go run .
```

This will start the interactive mode. Alternatively, you can execute commands directly:

```bash
go run . <command> <arguments>
```

For help with available commands, use the `help` command:

```bash
go run . help
```

or

```bash
go run . help <command>
```

Refer to the Go documentation for more details on the underlying `librobot` package.

## Commands

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

## Further Information

For more detailed information on the `librobot` package, refer to the Go documentation.

# Original Instructions:

We wish to create a simple Golang application which allows the use of the `librobot` simulator from a command line for testing purposes.

Use whatever Golang libraries you see fit.

### Part One

Create an interactive REPL type command line application which simulates a warehouse containing one or more robots, and allows issuing of tasks to these robots interactively.

The application should provide a prompt where the user is able to enter a task for a robot in string form.  The state of the simulated environment should be persisted between tasks being entered.

Provide relevant user documentation to allow the use of the application.

### Part Two

Add some kind of print out representation of the state of the simulation to the CLI application, which allows the user to see the simulation evolving in real time.
