package librobot

// This file performs unit tests on the librobot package to verify correct operation

import (
	"bytes"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestNewWarehouse checks if warehouse creation is successful.
func TestNewWarehouse(t *testing.T) {
	w := NewWarehouse()
	if w == nil {
		t.Fatal("NewWarehouse returned nil")
	}

	whImpl, ok := w.(*warehouseImpl) // Retrieve instance of implementation
	if !ok {
		t.Fatal("NewWarehouse did not return a *warehouseImpl")
	}
	if whImpl.robots == nil {
		t.Error("warehouseImpl.robots map not initialized")
	}
	if whImpl.mu == nil {
		t.Error("warehouseImpl.mu mutex not initialized")
	}
	// Check grid initialization (should be all empty strings)
	for y := uint(0); y <= GridSize; y++ {
		for x := uint(0); x <= GridSize; x++ {
			if whImpl.gridyx[y][x] != "" {
				t.Errorf("grid[%d][%d] not empty on initialization, got %q", y, x, whImpl.gridyx[y][x])
			}
		}
	}
}

// Test add robot functionality and lists robots added (two)
func TestRobot_AddRobot(t *testing.T) {
	t.Log("Starting TestRobot_AddRobot")
	test_warehouse := NewWarehouse()
	// Add robot no name
	test_robot, err := AddRobot(test_warehouse, 0, 0, "")
	if err != nil {
		t.Fatalf("Unexpected error when adding robot; error: %v", err)
	}
	t.Log("Robot added at (0,0)")

	if test_robot == nil {
		t.Fatal("Expected robot, retrieved nil")
	}

	// Check robot position
	if test_robot.CurrentState().X != 0 || test_robot.CurrentState().Y != 0 {
		t.Errorf("Robot 1 initial state incorrect: expected (0,0), got (%d,%d)", test_robot.CurrentState().X, test_robot.CurrentState().Y)
	}

	// Check robot list
	if len(test_warehouse.Robots()) != 1 {
		t.Errorf("Expected 1 robot in warehouse, got %d", len(test_warehouse.Robots()))
	}

	// Check robot is recorded correctly in grid
	whImpl := test_warehouse.(*warehouseImpl)
	if whImpl.gridyx[0][0] == "" {
		t.Error("Robot 1 not recorded in warehouse grid at (0,0)")
	}

	// Add a second robot, in occupied position
	// Test adding a robot at an occupied position
	robot2, err := AddRobot(test_warehouse, 0, 0, "R2")
	if err == nil {
		t.Errorf("AddRobot(w, 0, 0) with occupied position: expected error, got %v", err)
	}
	if robot2 != nil {
		t.Error("AddRobot returned non-nil robot for occupied position")
	}
	// Confirm failed to add robot
	if len(test_warehouse.Robots()) != 1 { // Still should be only 1 robot
		t.Errorf("Expected 1 robot after failed AddRobot, got %d", len(test_warehouse.Robots()))
	}

	// Test robot added to out of bounds condition
	// Test adding a robot out of bounds
	r3, err := AddRobot(test_warehouse, GridSize+1, 0, "R3")
	if err == nil {
		t.Errorf("AddRobot(w, GridSize+1, 0): expected error, got %v", err)
	}
	if r3 != nil {
		t.Error("AddRobot returned non-nil robot for out of bounds position")
	}

	// Test adding another robot at a valid, unoccupied position
	r4, err := AddRobot(test_warehouse, 5, 5, "R4")
	if err != nil {
		t.Fatalf("AddRobot(w, 5, 5) failed: %v", err)
	}
	if r4.CurrentState().X != 5 || r4.CurrentState().Y != 5 {
		t.Errorf("Robot 4 initial state incorrect: expected (5,5), got (%d,%d)", r4.CurrentState().X, r4.CurrentState().Y)
	}
	if len(test_warehouse.Robots()) != 2 {
		t.Errorf("Expected 2 robots in warehouse, got %d", len(test_warehouse.Robots()))
	}
	t.Log("Passed all add robot conditions")
}

// TestRobot_EnqueueTask Test adding tasks with commands to the queue of a robot
func TestRobot_EnqueueTask(t *testing.T) {
	t.Log("Starting TestRobot_EnqueueTask")
	test_warehouse := NewWarehouse()
	test_robot, _ := AddRobot(test_warehouse, 0, 0, "R1")

	t.Log("Setting up TestRobot_EnqueueTask")
	taskID, posCh, errCh := test_robot.EnqueueTask("N")

	select {
	case state := <-posCh:
		if state.X != 0 || state.Y != 1 {
			t.Errorf("Expected (0,1) after N, got (%d,%d)", state.X, state.Y)
		}
	case err := <-errCh:
		t.Fatalf("Task %s failed: %v", taskID, err)
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for N command completion")
	}

	// Ensure task completes
	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Task %s failed unexpectedly: %v", taskID, err)
		}
	case <-time.After(time.Second): // Small wait for channel to close
		// OK, likely finished
	}
	if test_robot.CurrentState().X != 0 || test_robot.CurrentState().Y != 1 {
		t.Errorf("Final state after N incorrect: expected (0,1), got (%d,%d)", test_robot.CurrentState().X, test_robot.CurrentState().Y)
	}

}

// TestRobot_CancelTask tests cancelling a task conditions such as in progress
func TestRobot_CancelTask(t *testing.T) {
	t.Log("Starting TestRobot_CancelTask")
	test_warehouse := NewWarehouse()
	r, err := AddRobot(test_warehouse, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	initialState := r.CurrentState()

	//  Test cancelling a task in progress
	longTaskCommands := "NNNNNNNNNN" // 10 commands
	taskID1, posCh1, errCh1 := r.EnqueueTask(longTaskCommands)

	// Wait for a few commands to execute, then cancel
	//time.Sleep(2 * CommandExecutionTime) // Let 2 commands run

	// Better, more deterministic
	// Wait for a few commands to execute deterministically
	const updatesToWait = 2
	var updatesReceived sync.WaitGroup
	updatesReceived.Add(updatesToWait)

	go func() {
		for range posCh1 {
			updatesReceived.Done()
		}
	}()

	// Wait until we know at least 2 commands have finished.
	// We'll use a timeout here just in case something is wrong with the task execution.
	waitTimeout := time.After(time.Duration(updatesToWait+1) * CommandExecutionTime)
	done := make(chan struct{})
	go func() {
		updatesReceived.Wait()
		close(done)
	}()

	select {
	case <-done:
		// We've received the expected number of updates, now we can cancel.
	case <-waitTimeout:
		t.Fatal("Timeout waiting for position updates before cancellation")
	}

	// Cancel the task
	err = r.CancelTask(taskID1)
	if err != nil {
		t.Fatalf("Failed to cancel task 1: %v", err)
	}

	// Verify task 1 aborts and returns cancellation error
	select {
	case finalErr := <-errCh1:
		if finalErr.Error() != "task cancelled" { // Check for specific cancellation error
			t.Errorf("Task 1: Expected 'task cancelled' error, got %v", finalErr)
		}
	case <-time.After(2 * CommandExecutionTime): // Wait a bit more for cancellation to propagate
		t.Fatal("Timeout waiting for Task 1 cancellation error")
	}

	// Ensure position channel is closed (no more updates)
	select {
	case _, ok := <-posCh1:
		if ok {
			t.Error("Position channel 1 is still open after cancellation")
		}
	default: // If nothing immediately, means it might be closed. Need to re-check after a brief moment
	}

	// Give the robot time to settle after cancellation before next task
	time.Sleep(CommandExecutionTime)
	robotStateAfterCancel := r.CurrentState()
	// It should have moved 2 units north from (0,0) before cancellation.
	if robotStateAfterCancel.X != initialState.X || robotStateAfterCancel.Y != initialState.Y+2 {
		t.Errorf("Robot state after cancellation unexpected. Expected (%d,%d), got (%d,%d)",
			initialState.X, initialState.Y+2, robotStateAfterCancel.X, robotStateAfterCancel.Y)
	}

	// Test cancelling a queued task (should not start)
	taskID2, posCh2, errCh2 := r.EnqueueTask("EEEE") // 4 commands

	// Cancel task 2 immediately before it has a chance to start
	err = r.CancelTask(taskID2)
	if err != nil {
		t.Fatalf("Failed to cancel task 2: %v", err)
	}

	// Wait much longer than the task would take
	select {
	case state := <-posCh2:
		t.Errorf("Task 2: Received unexpected position update for cancelled queued task: %+v", state)
	case finalErr := <-errCh2:
		if finalErr.Error() != "task cancelled" {
			t.Errorf("Task 2: Expected 'task cancelled' error, got %v", finalErr)
		}
	case <-time.After(5 * CommandExecutionTime): // Longer than 4 commands + overhead
		// This is good, means it likely never started or was quickly cancelled.
		t.Log("Task 2: Timeout reached, likely cancelled before execution.")
	}

	// Verify task 2's channels are closed (no activity)
	select {
	case _, ok := <-posCh2:
		if ok {
			t.Error("Position channel 2 is still open after cancellation")
		}
	default:
	}

	// Final robot position should remain the same as after task 1 cancellation
	if r.CurrentState() != robotStateAfterCancel {
		t.Errorf("Robot state changed after cancelling queued task. Expected %+v, got %+v", robotStateAfterCancel, r.CurrentState())
	}

	// 3. Test cancelling a non-existent task
	err = r.CancelTask("non-existent-task-id")
	if err == nil {
		t.Errorf("Expected error for non-existent task, got %v", err)
	}
}

// TestRobot_CancelTaskTwice Cancel task that has already been cancelled
func TestRobot_CancelTaskTwice(t *testing.T) {
	t.Log("Starting TestRobot_CancelTaskTwice")
	test_warehouse := NewWarehouse()
	r, err := AddRobot(test_warehouse, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	taskID, _, _ := r.EnqueueTask("N")

	err = r.CancelTask(taskID)
	if err != nil {
		t.Fatalf("Failed to cancel task: %v", err)
	}

	// Call cancel task twice
	err = r.CancelTask(taskID)
	if err == nil {
		t.Errorf("Expected error for cancelling same task twice, got %v", err)
	}
}

// TestRobot_GrabCrate Test robot picking crate function
func TestRobot_GrabCrate(t *testing.T) {
	t.Log("Starting TestRobot_GrabCrate")
	cw := NewCrateWarehouse()
	r, err := AddRobot(cw, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Add crate
	cw.AddCrate(0, 0)

	// Try to grab crate
	taskID, _, errCh := r.EnqueueTask("G")
	select {
	case finalErr := <-errCh:
		if finalErr != nil {
			t.Errorf("Error occured picking crate, got: %v on taskid %v", finalErr, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate pick command")
	}
}

// Drop crate and drop on existing crate
func TestRobot_DropCrate(t *testing.T) {
	t.Log("Starting TestRobot_DropCrate")
	cw := NewCrateWarehouse()
	r, err := AddRobot(cw, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Add crate
	cw.AddCrate(0, 0)

	// Try to grab crate
	taskID1, _, errCh1 := r.EnqueueTask("GN") // Grab crate, move forwards
	select {
	case finalErr := <-errCh1:
		if finalErr != nil {
			t.Errorf("Error occured picking crate, got: %v on taskid %v", finalErr, taskID1)
		}
	case <-time.After(4 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate pick command")
	}

	cw.AddCrate(0, 1)

	// Try to drop crate
	taskID2, _, errCh2 := r.EnqueueTask("D")
	select {
	case finalErr := <-errCh2:
		if finalErr != ErrCrateExists {
			t.Errorf("Unexpected error occured dropping crate, got: %v on taskid %v", finalErr, taskID2)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for incorrect crate drop command")
	}

	// Move forwards, drop crate; should be successful this time
	// Try to drop crate
	taskID3, _, errCh3 := r.EnqueueTask("ND")
	select {
	case finalErr := <-errCh3:
		if finalErr != nil {
			t.Errorf("Unexpected error occured dropping crate, got: %v on taskid %v", finalErr, taskID3)
		}
	case <-time.After(3 * CommandExecutionTime):
		t.Fatal("Timeout waiting for correct crate drop command")
	}

	// Check there is a crate at 0, 2
}

// TestRobot_CollisionDetection Test collision detection and correct errors are displayed
func TestRobot_CollisionDetection(t *testing.T) {
	t.Log("Starting TestRobot_CollisionDetection")
	w := NewWarehouse()
	r1, err := AddRobot(w, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	_, err = AddRobot(w, 1, 0, "R2")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Try to move robot to occupied position
	taskID, _, errCh := r1.EnqueueTask("E")
	select {
	case finalErr := <-errCh:
		if finalErr == nil {
			t.Errorf("Expected error moving robot to occupied position, got: %v on taskid %v", finalErr, taskID)
		}
		if finalErr != ErrPositionOccupied {
			t.Errorf("Unexpected error occured when moving to occupied position; got %v", finalErr)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate drop command")
	}
}

// TestRobot_NoListeners Test no listeners to robot
func TestRobot_NoListeners(t *testing.T) {
	t.Log("Starting TestRobot_NoListeners")
	w := NewWarehouse()
	r, err := AddRobot(w, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Enqueue task without listeners
	r.EnqueueTask("N")

	// Wait for a while to allow the task to complete
	time.Sleep(2 * CommandExecutionTime)
}

// TestCrateWarehouse_Robot Tests setting up and manipulating a crate warehouse with various scenarios
func TestCrateWarehouse_Robot(t *testing.T) {
	// create new Crate warehouse
	cw := NewCrateWarehouse()
	cwImpl := cw.(*warehouseImpl)
	// Add robot
	r1, err := AddRobot(cw, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Try to enqueue a crate command in a crate warehouse with no crates
	taskID, _, errCh := r1.EnqueueTask("G")
	select {
	case finalErr := <-errCh:
		if finalErr == nil {
			t.Errorf("Expected error, got: %v on taskID %v", finalErr, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for invalid command error")
	}
	// Try drop crate not exists
	taskID, _, errCh = r1.EnqueueTask("D")
	select {
	case finalErr := <-errCh:
		if finalErr == nil {
			t.Errorf("Expected error, got: %v on taskid %v", finalErr, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for invalid command error")
	}

	// Add crates
	cw.AddCrate(0, 0)

	// Now try pick crate
	taskID1, _, errCh1 := r1.EnqueueTask("G")
	// Wait for completion
	time.Sleep(2 * CommandExecutionTime)

	select {
	case finalErr := <-errCh1:
		if finalErr != nil {
			t.Errorf("Error occured picking crate, got: %v on taskid %v", finalErr, taskID1)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate pick command")
	}

	// Check crate no longer exists; should be false
	if cwImpl.cratesyx[0][0] {
		t.Fatalf("Error crate not picked; on task %v", taskID1)
	}
	// Check robot holding crate
	if !r1.CurrentState().HasCrate {
		t.Fatalf("Error crate not picked by robot; on task %v", taskID1)
	}

	// Add crate again
	cw.AddCrate(0, 0)
	// Try pick crate again; should have error
	taskID_repeat, _, errCh_repeat := r1.EnqueueTask("G")
	// Wait for completion
	time.Sleep(2 * CommandExecutionTime)
	select {
	case finalErr := <-errCh_repeat:
		if finalErr != ErrRobotHasCrate {
			t.Errorf("Error occured picking crate, got: %v on taskid %v", finalErr, taskID_repeat)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate pick command")
	}
	// Check crate still exists; should be true
	if !cwImpl.cratesyx[0][0] {
		t.Fatalf("Error crate not picked; on task %v", taskID1)
	}
	// Check robot holding crate
	if !r1.CurrentState().HasCrate {
		t.Fatalf("Error crate not picked by robot; on task %v", taskID1)
	}
	// remove crate we added
	cw.DelCrate(0, 0)

	// Now try drop crate
	taskID2, _, errCh2 := r1.EnqueueTask("D")
	// Wait for completion
	time.Sleep(2 * CommandExecutionTime)
	select {
	case finalErr := <-errCh2:
		if finalErr != nil {
			t.Errorf("Error occured dropping crate, got: %v on taskid %v", finalErr, taskID2)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for crate drop command")
	}
	// Check crate no longer exists; should be false
	if !cwImpl.cratesyx[0][0] {
		t.Fatalf("Error crate not dropped; on task %v", taskID1)
	}
	// Check robot holding crate
	if r1.CurrentState().HasCrate {
		t.Fatalf("Error crate not dropped by robot; on task %v", taskID1)
	}

}

// TestWarehouse_CrateCommands sets up crate warehouse and validates various crate manipulation commands
func TestWarehouse_CrateCommands(t *testing.T) {
	w := NewWarehouse()
	r, err := AddRobot(w, 0, 0, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	// Try to enqueue a crate command in a basic warehouse
	taskID, _, errCh := r.EnqueueTask("G")
	select {
	case finalErr := <-errCh:
		if finalErr == nil {
			t.Errorf("Expected error, got: %v on taskID %v", finalErr, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for invalid command error")
	}

	taskID, _, errCh = r.EnqueueTask("D")
	select {
	case finalErr := <-errCh:
		if finalErr == nil {
			t.Errorf("Expected error, got: %v on taskid %v", finalErr, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for invalid command error")
	}
}

// TestCrateManagement checks AddCrate and DelCrate functionality.
func TestCrateManagement(t *testing.T) {
	cw := NewCrateWarehouse()
	cwImpl := cw.(*warehouseImpl)

	// Test AddCrate
	err := cw.AddCrate(1, 1)
	if err != nil {
		t.Fatalf("AddCrate(1,1) failed: %v", err)
	}
	if !cwImpl.cratesyx[1][1] {
		t.Error("Crate not found at (1,1) after AddCrate")
	}

	// Test AddCrate to occupied spot
	err = cw.AddCrate(1, 1)
	if err != ErrCrateExists {
		t.Errorf("AddCrate(1,1) (occupied): Expected %v, got %v", ErrCrateExists, err)
	}

	// Test AddCrate out of bounds
	err = cw.AddCrate(GridSize+1, 1)
	if err != ErrCrateOutOfBounds {
		t.Errorf("AddCrate(out of bounds): Expected %v, got %v", ErrCrateOutOfBounds, err)
	}

	// Test DelCrate
	err = cw.DelCrate(1, 1)
	if err != nil {
		t.Fatalf("DelCrate(1,1) failed: %v", err)
	}
	if cwImpl.cratesyx[1][1] {
		t.Error("Crate found at (1,1) after DelCrate")
	}

	// Test DelCrate from empty spot
	err = cw.DelCrate(1, 1)
	if err != ErrCrateNotFound {
		t.Errorf("DelCrate(1,1) (empty): Expected %v, got %v", ErrCrateNotFound, err)
	}

	// Test DelCrate out of bounds
	err = cw.DelCrate(GridSize+1, 1)
	if err != ErrCrateOutOfBounds {
		t.Errorf("DelCrate(out of bounds): Expected %v, got %v", ErrCrateOutOfBounds, err)
	}

	// Test Add/Del Crate on a non-CrateWarehouse (should fail)
	w := NewWarehouse()
	err = w.(CrateWarehouse).AddCrate(1, 1) // Type assert to call method
	if err != ErrInvalidWarehouseType {
		t.Errorf("AddCrate on non-CrateWarehouse: Expected %v, got %v", ErrInvalidWarehouseType, err)
	}
	err = w.(CrateWarehouse).DelCrate(1, 1)
	if err != ErrInvalidWarehouseType {
		t.Errorf("DelCrate on non-CrateWarehouse: Expected %v, got %v", ErrInvalidWarehouseType, err)
	}
}

// TestInvalidCommands checks command parsing and ignores invalid commands
func TestInvalidCommands(t *testing.T) {
	w := NewWarehouse()
	r, err := AddRobot(w, 5, 5, "R1")
	if err != nil {
		t.Fatalf("Failed to add robot: %v", err)
	}

	initialState := r.CurrentState()

	taskID, posCh, errCh := r.EnqueueTask("NX") // N then an invalid command 'X'
	select {
	case state := <-posCh:
		// First command 'N' should succeed, robot moves to (5,6)
		if state.X != 5 || state.Y != 6 {
			t.Errorf("Expected (5,6) after 'N', got (%d,%d) on task %v", state.X, state.Y, taskID)
		}
	case <-time.After(2 * CommandExecutionTime):
		t.Fatal("Timeout waiting for first valid command")
	}

	select {
	case taskErr := <-errCh:
		expectedErrStr := "unknown command: X"
		if taskErr == nil || taskErr.Error() != expectedErrStr {
			t.Errorf("Expected error '%s', got '%v'", expectedErrStr, taskErr)
		}
	case <-time.After(2 * CommandExecutionTime): // Wait for error from second command
		t.Fatal("Timeout waiting for invalid command error")
	}

	// Robot should be at (5,6) (after 'N' and before 'X' aborted the task)
	finalState := r.CurrentState()
	if finalState.X != 5 || finalState.Y != 6 {
		t.Errorf("Robot state incorrect after invalid command abort. Expected (5,6), got (%d,%d)", finalState.X, finalState.Y)
	}
	if finalState == initialState {
		t.Error("Robot state should have changed after first valid command")
	}
}

// TestDiagonalRobot test basic robot function; add robot, add to invalid position etc
func TestDiagonalRobot(t *testing.T) {
	w := NewWarehouse()

	// Add robot
	_, err := AddDiagonalRobot(w, 5, 5, "")
	if err != nil {
		t.Fatalf("Failed to add diagonal robot: %v", err)
	}

	// Add robot occupied
	// Add robot
	_, err2 := AddDiagonalRobot(w, 5, 5, "")
	if err2 != ErrPositionOccupied {
		t.Fatalf("Unexpected error on position occupied: %v", err2)
	}

	// Add robot invalid
	_, err3 := AddDiagonalRobot(w, GridSize+2, GridSize+2, "Error3")
	if err3 != ErrOutOfBounds {
		t.Fatalf("Unexpected error on out of bounds: %v", err3)
	}
}

// TestDiagonalMovementComplexCommands Test complex diagonal commands
func TestDiagonalMovementComplexCommands(t *testing.T) {
	w := NewWarehouse()

	r, err := AddDiagonalRobot(w, 5, 5, "")
	if err != nil {
		t.Fatalf("Failed to add diagonal robot: %v", err)
	}

	// Command sequence: "N E E N W W"
	// Expected movement: NE (to 6,6), EN (to 7,7), W (to 6,7), W (to 5,7), N (to 5,7)
	// The "N" and "W" are a pair, but the "W" and "W" are not.
	// The "E" and "E" are not a pair.
	// The "E" and "N" are a pair.

	taskID, posCh, errCh := r.EnqueueTask("N E E N W W N")

	expectedStates := []RobotState{
		{X: 6, Y: 6}, // NE
		{X: 7, Y: 7}, // EN
		{X: 6, Y: 7}, // W
		{X: 5, Y: 8}, // WN
	}

	for i, expected := range expectedStates {
		select {
		case state := <-posCh:
			if state.X != expected.X || state.Y != expected.Y {
				t.Errorf("Step %d: Expected (%d,%d), got (%d,%d) on task %v", i, expected.X, expected.Y, state.X, state.Y, taskID)
			}
		case err := <-errCh:
			t.Fatalf("Task failed on step %d: %v", i, err)
		case <-time.After(2 * CommandExecutionTime):
			t.Fatalf("Timeout waiting for step %d", i)
		}
	}
}

// TestTableDiagonalMovementComplexCommands Table test for a number of different diagonal movement commands and cases
func TestTableDiagonalMovementComplexCommands(t *testing.T) {
	// ... test setup (create warehouse and diagonal robot)

	// Create tables for testing states
	testCases := []struct {
		name               string
		commands           string
		initialX, initialY uint
		expectedStates     []RobotState
		expectError        error
	}{
		{
			name:     "NE-EN-WW",
			commands: "NEENWW",
			initialX: 5, initialY: 5,
			expectedStates: []RobotState{
				{X: 6, Y: 6}, // NE
				{X: 7, Y: 7}, // EN (diagonal)
				{X: 6, Y: 7}, // W
				{X: 5, Y: 7}, // W
			},
			expectError: nil,
		},
		{
			name:     "WNW (fused)",
			commands: "WNW",
			initialX: 5, initialY: 5,
			expectedStates: []RobotState{
				{X: 4, Y: 6}, // WN (diagonal)
				{X: 3, Y: 6}, // W
			},
			expectError: nil,
		},
		{
			name:     "SSWSE (fused)",
			commands: "SSWSE",
			initialX: 5, initialY: 5,
			expectedStates: []RobotState{
				{X: 5, Y: 4}, // S
				{X: 4, Y: 3}, // SW  (diagonal)
				{X: 5, Y: 2}, // SE  (diagonal)
			},
			expectError: nil,
		},
		{
			name:     "N E E N W W (with spaces)",
			commands: "N E E N W W",
			initialX: 5, initialY: 5,
			expectedStates: []RobotState{
				{X: 6, Y: 6}, // NE (fused)
				{X: 7, Y: 7}, // EN (fused)
				{X: 6, Y: 7}, // W
				{X: 5, Y: 7}, // W
			},
			expectError: nil,
		},
		{
			name:     "E E (no fusion)",
			commands: "EE",
			initialX: 5, initialY: 5,
			expectedStates: []RobotState{
				{X: 6, Y: 5}, // E
				{X: 7, Y: 5}, // E
			},
			expectError: nil,
		},
		{
			name:     "Diagonal out of bounds (SW from 0,0)",
			commands: "SW",
			initialX: 0, initialY: 0,
			expectedStates: []RobotState{},
			expectError:    ErrOutOfBounds,
		},
	}

	// Loop over the test cases.
	for _, tc := range testCases {
		// Use t.Run() for cleaner output and isolation.
		t.Run(tc.name, func(t *testing.T) {
			// Setup a fresh environment for each test case.
			w := NewWarehouse()
			r, err := AddDiagonalRobot(w, tc.initialX, tc.initialY, "Robot1")
			if err != nil {
				t.Fatalf("Failed to add diagonal robot: %v", err)
			}

			_, posCh, errCh := r.EnqueueTask(tc.commands)

			// Track the robot's state through the position channel.
			var receivedStates []RobotState
			for state := range posCh {
				receivedStates = append(receivedStates, state)
			}

			// Check for errors
			var finalErr error
			select {
			case err := <-errCh:
				finalErr = err
			default:
				// No error, continue
			}

			// Validate the final error.
			if finalErr != tc.expectError {
				t.Fatalf("Expected error %v, got %v", tc.expectError, finalErr)
			}

			// Validate the final state sequence if no error was expected.
			if tc.expectError == nil {
				if len(receivedStates) != len(tc.expectedStates) {
					t.Fatalf("Expected %d states, got %d", len(tc.expectedStates), len(receivedStates))
				}
				for i := range receivedStates {
					if receivedStates[i] != tc.expectedStates[i] {
						t.Errorf("Step %d: Expected state %+v, got %+v", i, tc.expectedStates[i], receivedStates[i])
					}
				}
			}
		})
	}
}

// TestRender tests the output of the render engine
func TestRender_PrintsCorrectWarehouseView(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	stdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Setup warehouse
	wImpl := NewCrateWarehouse()
	wImpl.AddCrate(1, 1)
	wImpl.AddCrate(2, 2)
	wImpl.AddCrate(3, 3)
	robot, err := AddRobot(wImpl, 1, 1, "R1")
	if err != nil {
		t.Fatalf("Error adding robot: %v\n", err)
		return
	}

	robot2, err2 := AddRobot(wImpl, 3, 3, "R2")
	if err2 != nil {
		t.Fatalf("Error adding robot: %v\n", err)
		return
	}

	robot_map := make(map[string]Robot) // Map of robots to user defined robot IDs
	// Add to our own map of robot ids
	robot_map["R1"] = robot
	robot_map["R2"] = robot2

	robot2.EnqueueTask("G")
	time.Sleep(1 * CommandExecutionTime)

	// Call render
	Render(wImpl, robot_map)

	// Finish capturing output
	w.Close()
	os.Stdout = stdout
	buf.ReadFrom(r)

	// Extract the printed string
	output := buf.String()

	// Basic assertions
	if !strings.Contains(output, "R1_") {
		t.Errorf("Expected robot with crate under to render as 'r1_', got:\n%s", output)
	}
	if !strings.Contains(output, "R2*") {
		t.Errorf("Expected robot with crate under to render as 'r2*', got:\n%s", output)
	}
	if !strings.Contains(output, "[C]") {
		t.Errorf("Expected crate to render as [C], got:\n%s", output)
	}
}
