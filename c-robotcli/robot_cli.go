package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"robot_challenge/b-librobot/librobot"

	"github.com/spf13/cobra"
)

// Global variables to be used by all commands
// TODO: explore if there are ways to do this without global variable in Go
var (
	warehouse      librobot.CrateWarehouse
	done           chan bool
	simulationTick = 200 * time.Millisecond
	robot_map      map[string]librobot.Robot
	viewIsRunning  bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "robot-cli",
	Short: "An interactive robot warehouse simulator",
	Long: `A command-line application that simulates a warehouse with robots
and allows you to issue tasks to them. The simulation state is displayed
in real-time.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Robot CLI invoked. Use the available commands to control the robot.")
	},
}

// addRobotCmd represents the add_robot command
var addRobotCmd = &cobra.Command{
	Use:   "add_robot [id] [x] [y]",
	Short: "Add a new robot to the warehouse",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		id := args[0]
		x, errX := strconv.Atoi(args[1])
		y, errY := strconv.Atoi(args[2])

		if errX != nil || errY != nil {
			fmt.Println("Error: Invalid coordinates. Please use integers.")
			return
		}
		robot, err := librobot.AddRobot(warehouse, uint(x), uint(y))
		if err != nil {
			fmt.Printf("Error adding robot: %v %v\n", id, err)
			return
		}

		// Add to our own map of robot ids
		robot_map[id] = robot
		fmt.Printf("Added robot '%s' at (%d, %d).\n", id, x, y)
	},
}

// addTaskCmd represents the add_task command
var addTaskCmd = &cobra.Command{
	Use:   "add_task [robot_id] [commands]",
	Short: "Enqueue a task for a robot",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		robotID := args[0]
		commands := strings.Join(args[1:], "")

		// Get robot from map
		robot, ok := robot_map[robotID]

		if !ok {
			fmt.Printf("Error: Robot with ID '%s' not found.\n", robotID)
			return
		}

		taskID, _, errChan := robot.EnqueueTask(commands)
		fmt.Printf("Task '%s' enqueued for robot '%s'.\n", taskID, robotID)

		// Listen for task completion/errors in a non-blocking way
		go func() {
			for err, ok := <-errChan; ok; err, ok = <-errChan {
				if err != nil {
					fmt.Printf("Task '%s' for robot '%s' failed: %v\n", taskID, robotID, err)
				}
			}
		}()
	},
}

// addCrateCmd represents the add_crate command
var addCrateCmd = &cobra.Command{
	Use:   "add_crate [x] [y]",
	Short: "Add a new crate to the warehouse",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		x, errX := strconv.Atoi(args[0])
		y, errY := strconv.Atoi(args[1])
		if errX != nil || errY != nil {
			fmt.Println("Error: Invalid coordinates. Please use integers.")
			return
		}

		if err := warehouse.AddCrate(uint(x), uint(y)); err != nil {
			fmt.Printf("Error adding crate: %v\n", err)
			return
		}
		fmt.Printf("Crate added at (%d, %d).\n", x, y)
	},
}

// delCrateCmd represents the del_crate command
var delCrateCmd = &cobra.Command{
	Use:   "del_crate [x] [y]",
	Short: "Deletes a crate from the warehouse",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		x, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid x coordinate: %v", err)
		}
		y, err := strconv.ParseUint(args[1], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid y coordinate: %v", err)
		}
		return warehouse.DelCrate(uint(x), uint(y))
	},
}

// cancelTaskCmd represents the cancel_task command
var cancelTaskCmd = &cobra.Command{
	Use:   "cancel_task [robot_id] [task_id]",
	Short: "Cancel a running task for a robot",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		robotID := args[0]
		taskID := args[1]

		// Get robot from map
		robot, ok := robot_map[robotID]

		if !ok {
			fmt.Printf("Error: Robot with ID '%s' not found.\n", robotID)
			return
		}

		if err := robot.CancelTask(taskID); err != nil {
			fmt.Printf("Error canceling task: %v\n", err)
			return
		}
		fmt.Printf("Task '%s' for robot '%s' canceled.\n", taskID, robotID)
	},
}

// viewCmd starts the visualization in a separate goroutine
var viewCmd = &cobra.Command{
	Use:   "view",
	Short: "Shows a real-time ASCII view of the warehouse",
	Run: func(cmd *cobra.Command, args []string) {
		if viewIsRunning {
			fmt.Println("View is already running. Use 'stop_view' to stop it.")
			return
		}

		// Re-initialize the done channel and set the running flag
		done = make(chan bool)
		viewIsRunning = true

		// Clear the screen once to provide a clean canvas for the view.
		librobot.ClearScreen()

		go func() {
			ticker := time.NewTicker(simulationTick)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					// Reset cursor to top left and refresh rendering
					fmt.Print("\033[H")
					librobot.Render(warehouse)
				case <-done:
					fmt.Println("Stopping view...")
					return
				}
			}
		}()
		fmt.Println("View started. Use 'stop_view' to halt rendering.")
	},
}

// stopViewCmd stops the visualization goroutine
var stopViewCmd = &cobra.Command{
	Use:   "stop_view",
	Short: "Stops the real-time ASCII view",
	Run: func(cmd *cobra.Command, args []string) {
		if !viewIsRunning {
			fmt.Println("View is not running.")
			return
		}
		close(done)
		viewIsRunning = false
	},
}

// init function to set up Cobra commands
func init() {
	rootCmd.AddCommand(addRobotCmd)
	rootCmd.AddCommand(addTaskCmd)
	rootCmd.AddCommand(addCrateCmd)
	rootCmd.AddCommand(delCrateCmd)
	rootCmd.AddCommand(cancelTaskCmd)
	rootCmd.AddCommand(viewCmd)
	rootCmd.AddCommand(stopViewCmd)
}

func main() {
	// Create empty warehouse for interaction
	warehouse = librobot.NewCrateWarehouse()
	done = make(chan bool)
	robot_map = make(map[string]librobot.Robot) // Map of robots to user defined robot IDs

	// Check if any command-line arguments were provided.
	// This determines if running in interactive mode.
	if len(os.Args) > 1 {
		// Execute the command and exit.
		if err := rootCmd.Execute(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	// Interactive mode
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Interactive Robot CLI. Type 'exit' to quit.")
	fmt.Println("Use 'help' to see available commands.")
	fmt.Println("---")

	for {
		// If the view is running, we need to move the cursor to a new line
		// for the prompt, to prevent it from being overwritten.
		if viewIsRunning {
			// Calculate the row for the prompt: GridSize + header (4 lines)
			promptRow := librobot.GridSize + 4
			// Move cursor to the calculated row, column 0, and clear the line
			fmt.Printf("\033[%d;0H\033[K", promptRow)
		}

		// Print the prompt
		fmt.Print("> ")

		// Read the user's input
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		// Clean up the input string
		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// Handle the "exit" command
		if strings.ToLower(input) == "exit" {
			// Signal the view to stop before exiting
			if viewIsRunning {
				close(done)
			}
			fmt.Println("Exiting interactive CLI. Goodbye!")
			return
		}

		// Split the input into arguments
		args := strings.Split(input, " ")

		// Pass the arguments to the root command to simulate command-line execution
		rootCmd.SetArgs(args)

		// Execute the command and handle any errors.
		// Cobra's Execute() method can exit the program, so we need to
		// handle its execution carefully within the loop.
		if err := rootCmd.Execute(); err != nil {
			// Print the error but don't exit the program
			fmt.Println(err)
		}
	}

}
