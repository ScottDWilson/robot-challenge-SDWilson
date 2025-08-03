package librobot_test

import (
	"fmt"
	"log"
	"time"

	"robot_challenge/b-librobot/librobot"
)

// Example basic usage of the library
func Example_basicPhase1() {
	log.Println("Running example of basic librobot usage")
	// Create an empty warehouse
	warehouse1 := librobot.NewWarehouse()
	// Add 1 robot to warehouse
	robot1, err := librobot.AddRobot(warehouse1, 0, 0, "R1")
	if err != nil {
		log.Printf("Unexpected error when adding robot; error: %v \n", err)
	}
	// Send robot commands
	command_string := "N N E E N E S" // Robot starting at 0,0 will end up at position 2,3
	taskID, _, _ := robot1.EnqueueTask(command_string)

	// Wait for completion...
	log.Println("Robot in action... waiting for completion...")
	time.Sleep(8 * librobot.CommandExecutionTime) // 8* wait for robust; should be complete

	// Check robot position
	robot1_state := robot1.CurrentState()
	if robot1_state.X != 3 || robot1_state.Y != 2 {
		log.Fatalf("Robot failed to move to position on task %v", taskID)
	}

	log.Println("Basic Usage Example completed successfully")
	fmt.Printf("Robot Position %v : %v", robot1_state.X, robot1_state.Y)

	// Output:Robot Position 3 : 2
}

// Example_mutiple Provides an example implementation of multiple warehouses operating with multiple robots
func Example_mutiple() {
	warehouse1 := librobot.NewWarehouse()
	warehouse2 := librobot.NewWarehouse()

	// Add 2 robot to warehouse1
	robot1, err1 := librobot.AddRobot(warehouse1, 0, 0, "R1")
	robot2, err2 := librobot.AddRobot(warehouse1, 5, 5, "R2")
	if err1 != nil || err2 != nil {
		log.Printf("Unexpected error when adding robot; error: %v, %v \n", err1, err2)
	}
	// Send robot commands
	command_string := "N N E E N E S"                  // Robot starting at 0,0 will end up at position
	taskID, _, _ := robot1.EnqueueTask(command_string) // End on 3,2

	// Wait short time
	// Add more tasks to robot 1 during operation
	time.Sleep(2 * librobot.CommandExecutionTime)
	taskID1, _, _ := robot1.EnqueueTask("N N ") // End on 3, 4

	// Wait short time
	time.Sleep(2 * librobot.CommandExecutionTime)
	taskID2, _, _ := robot2.EnqueueTask(command_string) // End on 8,7

	// Wait short time
	time.Sleep(2 * librobot.CommandExecutionTime)
	// Add robot to second warehouse
	robot3, err3 := librobot.AddRobot(warehouse2, 9, 9, "R3")
	if err3 != nil {
		log.Printf("Unexpected error when adding robot; error: %v \n", err3)
	}
	taskID3, _, _ := robot3.EnqueueTask("S S W W S W") // End on 7,7

	// Wait for completion...
	log.Println("Robot in action... waiting for completion...")
	time.Sleep(7 * librobot.CommandExecutionTime) // 8* wait for robust; should be complete

	// Check robot positions
	robot1_state := robot1.CurrentState()
	if robot1_state.X != 3 || robot1_state.Y != 4 {
		log.Fatalf("Robot 1 failed to move to position on tasks %v %v \n", taskID, taskID1)
	}

	// Check robot positions
	robot2_state := robot2.CurrentState()
	if robot2_state.X != 8 || robot2_state.Y != 7 {
		log.Fatalf("Robot 2 failed to move to position on task %v \n", taskID2)
	}

	// Check robot positions
	robot3_state := robot3.CurrentState()
	if robot3_state.X != 6 || robot3_state.Y != 6 {
		log.Printf("robot3_state.X Y %v %v", robot3_state.X, robot3_state.Y)
		log.Fatalf("Robot 3 failed to move to position on task %v \n", taskID3)
	}

	// Check Warehouse Robots
	if len(warehouse1.Robots()) != 2 {
		log.Fatalf("Warehouse 1 has incorrect number of robots")
	}
	if len(warehouse2.Robots()) != 1 {
		log.Fatalf("Warehouse 2 has incorrect number of robots")
	}

	log.Println("Example completed successfully")
	fmt.Printf("Robot 1 End Position %v : %v\n", robot1_state.X, robot1_state.Y)
	fmt.Printf("Robot 2 End Position %v : %v\n", robot2_state.X, robot2_state.Y)
	fmt.Printf("Robot 3 End Position %v : %v\n", robot3_state.X, robot3_state.Y)

	// Output:
	// Robot 1 End Position 3 : 4
	// Robot 2 End Position 8 : 7
	// Robot 3 End Position 6 : 6
}

// Example_Phase2 Shows an implementation of Phase 2 where a robot can pick up and drop crates
func Example_phase2() {
	warehouse1 := librobot.NewCrateWarehouse()

	// Add robot to warehouse1
	robot1, err1 := librobot.AddRobot(warehouse1, 0, 0, "R1")
	if err1 != nil {
		log.Printf("Unexpected error when adding robot; error: %v \n", err1)
	}

	warehouse1.AddCrate(1, 1)

	// Move Robot to position and pick up crate
	robot1.EnqueueTask("NEG")

	// Wait for completion
	time.Sleep(3 * librobot.CommandExecutionTime)

	robot1_state := robot1.CurrentState()

	log.Println("Example Phase 2 completed successfully")
	fmt.Printf("Robot 2 End State %v : %v Crate: %v \n", robot1_state.X, robot1_state.Y, robot1_state.HasCrate)

	// Output:
	// Robot 2 End State 1 : 1 Crate: true

}

// Example_Phase3 Provides an example of a robot moving diagonally
func Example_phase3() {
	w := librobot.NewWarehouse()

	r, err := librobot.AddDiagonalRobot(w, 5, 5, "")
	if err != nil {
		log.Printf("Failed to add diagonal robot: %v", err)
	}

	// Command sequence: "N E E N W W"
	// Expected movement: NE (to 6,6), EN (to 7,7), W (to 6,7), W (to 5,7), N (to 5,7)
	// The "N" and "W" are a pair, but the "W" and "W" are not.
	// The "E" and "E" are not a pair.
	// The "E" and "N" are a pair.

	taskID, posCh, errCh := r.EnqueueTask("N E E N W W N")

	expectedStates := []librobot.RobotState{
		{X: 6, Y: 6}, // NE
		{X: 7, Y: 7}, // EN
		{X: 6, Y: 7}, // W
		{X: 5, Y: 8}, // WN
	}

	for i, expected := range expectedStates {
		select {
		case state := <-posCh:
			if state.X != expected.X || state.Y != expected.Y {
				log.Printf("Step %d: Expected (%d,%d), got (%d,%d) on task %v", i, expected.X, expected.Y, state.X, state.Y, taskID)
			}
		case err := <-errCh:
			log.Printf("Task failed on step %d: %v", i, err)
		case <-time.After(2 * librobot.CommandExecutionTime):
			log.Printf("Timeout waiting for step %d", i)
		}
	}

	robot_state := r.CurrentState()

	log.Println("Example Phase 3 completed successfully")
	fmt.Printf("Robot Phase 3 End State %v : %v Crate: %v \n", robot_state.X, robot_state.Y, robot_state.HasCrate)

	// Output:
	// Robot Phase 3 End State 5 : 8 Crate: false
}
