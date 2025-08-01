package librobot

// This file performs unit tests on the librobot package to verify correct operation

import (
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
	test_robot, err := AddRobot(test_warehouse, 0, 0)
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
	robot2, err := AddRobot(test_warehouse, 0, 0)
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
	r3, err := AddRobot(test_warehouse, GridSize+1, 0)
	if err == nil {
		t.Errorf("AddRobot(w, GridSize+1, 0): expected error, got %v", err)
	}
	if r3 != nil {
		t.Error("AddRobot returned non-nil robot for out of bounds position")
	}

	// Test adding another robot at a valid, unoccupied position
	r4, err := AddRobot(test_warehouse, 5, 5)
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

func TestRobot_EnqueueTask(t *testing.T) {
	t.Log("Starting TestRobot_EnqueueTask")
	test_warehouse := NewWarehouse()
	test_robot, _ := AddRobot(test_warehouse, 0, 0)

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

func TestRobot_CancelTask(t *testing.T) {
	t.Log("Starting TestRobot_CancelTask")
	test_warehouse := NewWarehouse()
	r, err := AddRobot(test_warehouse, 0, 0)
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
