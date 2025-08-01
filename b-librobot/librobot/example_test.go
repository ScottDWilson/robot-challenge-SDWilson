package librobot_test

import (
	"fmt"
	"log"
	"time"

	"librobot"
)

// Example basic usage of the library
func Example_basicPhase1() {
	log.Println("Running example of basic librobot usage")
	// Create an empty warehouse
	warehouse1 := librobot.NewWarehouse()
	// Add 1 robot to warehouse
	robot1, err := librobot.AddRobot(warehouse1, 0, 0)
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

func Example_mutiple() {
	warehouse1 := librobot.NewWarehouse()
	warehouse2 := librobot.NewWarehouse()

	// Add 2 robot to warehouse1
	robot1, err1 := librobot.AddRobot(warehouse1, 0, 0)
	robot2, err2 := librobot.AddRobot(warehouse1, 5, 5)
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
	robot3, err2 := librobot.AddRobot(warehouse2, 10, 10)
	taskID3, _, _ := robot3.EnqueueTask("S S W W S W") // End on 7,7

	// Wait for completion...
	log.Println("Robot in action... waiting for completion...")
	time.Sleep(8 * librobot.CommandExecutionTime) // 8* wait for robust; should be complete

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
	if robot3_state.X != 7 || robot3_state.Y != 7 {
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
	// Robot 3 End Position 7 : 7
}
