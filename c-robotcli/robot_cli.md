Robot Simulator CLI User Guide
This is an interactive command-line application for simulating a robot warehouse, now built with the Cobra framework. The application provides a structured command-line interface for managing robots and their tasks, with a real-time visualization of the simulation state.

Quick Start
Create a project directory, for example robot-cli.

Inside this directory, create a subdirectory librobot and place the librobot/librobot.go file there.

Place the main.go file in the main robot-sim directory.

Initialize the Go module and get the Cobra dependency:

go mod init robot-cli
go get github.com/spf13/cobra

Compile and run the application from your terminal:

go run .

The application will start, and you can now use the commands below.

Available Commands
add_robot [id] [x] [y]: Adds a new robot to the warehouse.

[id]: A unique identifier for the robot (e.g., robot_1).

[x]: The initial X coordinate (0-19).

[y]: The initial Y coordinate (0-19).

Example: robot-cli add_robot R2D2 5 10

add_task [robot_id] [commands]: Enqueues a task for a robot. The task is a string of single-character commands that the robot executes sequentially.

[robot_id]: The ID of the robot.

[commands]: A string of commands. The available commands are:

N: Move North (up)

E: Move East (right)

S: Move South (down)

W: Move West (left)

G: Pickup a crate at the current location.

D: Drop a crate at the current location.

Example:  robot-cli add_task R2D2 NNNWWWPD

add_crate [x] [y]: Adds a stationary crate to the warehouse at a specific location.

[x]: The X coordinate (0-19).

[y]: The Y coordinate (0-19).

Example:  robot-cli add_crate 5 7

cancel_task [robot_id] [task_id]: Cancels a running task.

[robot_id]: The ID of the robot with the task.

[task_id]: The unique ID returned when the task was enqueued.

Example:  robot-cli cancel_task R2D2 1678881234567890

help: Displays the help message for the root command or any subcommand.

Example: robot-cli help add_robot

Ctrl+C: Shuts down the application.

Simulation Visualization
The application provides a real-time, text-based grid to show the state of the warehouse.

Grid: The warehouse is a 10x10 grid.

Robots: Robots are represented by the letter R.

A blue R indicates a robot that is not carrying a crate.

A yellow R indicates a robot that is currently holding a crate.

Crates: Crates are represented by the letter C.

A green C indicates a crate on the floor.

State: Below the grid, the current state of each robot and the location of all crates are displayed.