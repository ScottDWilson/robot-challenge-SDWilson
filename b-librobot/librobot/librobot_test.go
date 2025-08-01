package librobot

// This file performs unit tests on the librobot package to verify correct operation

import (
	"testing"
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

}
