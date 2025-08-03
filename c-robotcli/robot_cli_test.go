package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"robot_challenge/b-librobot/librobot"
)

// setupTest initializes a new warehouse and robot_map for a fresh test.
func setupTest() {
	warehouse = librobot.NewCrateWarehouse()
	robot_map = make(map[string]librobot.Robot)
	// We do not start the view by default
	viewIsRunning = false
}

// captureOutput redirects stdout and stderr to a buffer and returns a function
// that restores them and returns the captured output.
func captureOutput() func() string {
	var buf bytes.Buffer
	r, w, _ := os.Pipe()
	stdout := os.Stdout
	stderr := os.Stderr
	os.Stdout = w
	os.Stderr = w

	return func() string {
		w.Close()
		os.Stdout = stdout
		os.Stderr = stderr
		io.Copy(&buf, r)
		r.Close()
		return buf.String()
	}
}

// TestAddRobot tests the "add_robot" command.
func TestAddRobot(t *testing.T) {
	setupTest()
	defer setupTest() // Clean up after the test

	// Capture the output to verify the command's success message.
	restoreOutput := captureOutput()
	defer restoreOutput()

	// Set the arguments for the "add_robot" command.
	RootCmd.SetArgs([]string{"add_robot", "r1", "1", "1"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_robot command failed: %v", err)
	}

	// Verify the robot was added to the warehouse.
	robots := warehouse.Robots()
	if len(robots) != 1 {
		t.Fatalf("Expected 1 robot, but found %d", len(robots))
	}

	// Check the robot's state
	state := robots[0].CurrentState()
	if state.X != 1 || state.Y != 1 {
		t.Fatalf("Expected robot at (1, 1), got (%d, %d)", state.X, state.Y)
	}

	// Check the printed message.
	output := restoreOutput()
	expectedOutput := "Added robot 'r1' at (1, 1)."
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}

	time.Sleep(1 * librobot.CommandExecutionTime)

	// Capture the output to verify the command's success message.
	restoreOutput = captureOutput()
	defer restoreOutput()

	// Test add invalid robot
	RootCmd.SetArgs([]string{"add_robot", "r1", "a", "b"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_robot command failed: %v", err)
	}

	output2 := restoreOutput()
	expectedOutput2 := "Error: Invalid coordinates"
	if !strings.Contains(output2, expectedOutput2) {
		t.Errorf("Expected output to contain %q, but got:\n%q", expectedOutput2, output2)
	} else {
		t.Logf("Output matched expected substring: %q", expectedOutput2)
	}

	// Capture the output to verify the command's success message.
	restoreOutput = captureOutput()
	defer restoreOutput()

	// Test add robot at existing position
	// Test add invalid robot
	RootCmd.SetArgs([]string{"add_robot", "r2", "1", "1"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_robot command failed: %v", err)
	}

	output3 := restoreOutput()
	expectedOutput3 := "Error adding robot:"
	if !strings.Contains(output3, expectedOutput3) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput3, output3)
	}

}

// TestAddDiagRobot tests the "add_diag_robot" command.
func TestAddDiagRobot(t *testing.T) {
	setupTest()
	defer setupTest() // Clean up after the test

	// Capture the output to verify the command's success message.
	restoreOutput := captureOutput()
	defer restoreOutput()

	// Set the arguments for the "add_robot" command.
	RootCmd.SetArgs([]string{"add_diag_robot", "r1", "1", "1"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_diag_robot command failed: %v", err)
	}

	// Verify the robot was added to the warehouse.
	robots := warehouse.Robots()
	if len(robots) != 1 {
		t.Fatalf("Expected 1 robot, but found %d", len(robots))
	}

	// Check the robot's state
	state := robots[0].CurrentState()
	if state.X != 1 || state.Y != 1 {
		t.Fatalf("Expected robot at (1, 1), got (%d, %d)", state.X, state.Y)
	}

	// Check the printed message.
	output := restoreOutput()
	expectedOutput := "Added robot 'r1' at (1, 1)."
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}

	time.Sleep(1 * librobot.CommandExecutionTime)

	// Capture the output to verify the command's success message.
	restoreOutput = captureOutput()
	defer restoreOutput()

	// Test add invalid robot
	RootCmd.SetArgs([]string{"add_diag_robot", "r1", "a", "b"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_diag_robot command failed: %v", err)
	}

	output2 := restoreOutput()
	expectedOutput2 := "Error: Invalid coordinates"
	if !strings.Contains(output2, expectedOutput2) {
		t.Errorf("Expected output to contain %q, but got:\n%q", expectedOutput2, output2)
	} else {
		t.Logf("Output matched expected substring: %q", expectedOutput2)
	}

	// Capture the output to verify the command's success message.
	restoreOutput = captureOutput()
	defer restoreOutput()

	// Test add robot at existing position
	// Test add invalid robot
	RootCmd.SetArgs([]string{"add_diag_robot", "r2", "1", "1"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_diag_robot command failed: %v", err)
	}

	output3 := restoreOutput()
	expectedOutput3 := "Error adding robot:"
	if !strings.Contains(output3, expectedOutput3) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput3, output3)
	}

}

// TestAddCrate tests the "add_crate" command.
func TestAddCrate(t *testing.T) {
	setupTest()
	defer setupTest()

	restoreOutput := captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"add_crate", "2", "3"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_crate command failed: %v", err)
	}

	// The `librobot` package doesn't have a public function to check for crates,
	// so we'll check the output message for now. A better approach would be to
	// add a public `HasCrate` function to the `librobot` interface.
	output := restoreOutput()
	expectedOutput := "Crate added at (2, 3)."
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}

	// Add invalid crate
	restoreOutput = captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"add_crate", "20", "30"})

	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_crate command failed: %v", err)
	}

	output = restoreOutput()
	expectedOutput = "Error adding crate:"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}
}

// TestDelCrate tests the "del_crate" command.
func TestDelCrate(t *testing.T) {
	setupTest()
	defer setupTest()

	// First, add a crate to be deleted.
	RootCmd.SetArgs([]string{"add_crate", "2", "3"})
	RootCmd.Execute()

	restoreOutput := captureOutput()
	defer restoreOutput()

	// Now, delete the crate.
	RootCmd.SetArgs([]string{"del_crate", "2", "3"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("del-crate command failed: %v", err)
	}

	output := restoreOutput()
	// The `del-crate` command in the current implementation does not print a
	// success message, so we just check for no errors.
	if len(output) > 0 {
		t.Errorf("Expected no output on success, but got:\n%s", output)
	}

	// Delete invalid crate X
	restoreOutput = captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"del_crate", "20", "3"})

	if err := RootCmd.Execute(); err == nil {
		t.Fatalf("Unexpected success; del_crate command: %v", err)
	}

	output = restoreOutput()
	expectedOutput := "error: crate out of bounds"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}

	// Delete invalid crate Y
	restoreOutput = captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"del_crate", "2", "30"})

	if err := RootCmd.Execute(); err == nil {
		t.Fatalf("Unexpected Success on del_create command: %v", err)
	}

	output = restoreOutput()
	expectedOutput = "error: crate out of bounds"
	if !strings.Contains(output, expectedOutput) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutput, output)
	}
}

// TestAddTask tests the "add_task" command.
func TestAddTask(t *testing.T) {
	setupTest()
	defer setupTest()

	// First, add a robot to assign a task to.
	RootCmd.SetArgs([]string{"add_robot", "r1", "1", "1"})
	RootCmd.Execute()

	restoreOutput := captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"add_task", "r1", "N", "E", "N"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_task command failed: %v", err)
	}

	output := restoreOutput()
	expectedOutputPrefix := "Task '"
	if !strings.Contains(output, expectedOutputPrefix) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutputPrefix, output)
	}

	// Test Task for non-existent robot
	restoreOutput = captureOutput()
	defer restoreOutput()

	RootCmd.SetArgs([]string{"add_task", "r8", "N", "E", "N"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("add_task command failed: %v", err)
	}

	output = restoreOutput()
	expectedOutputPrefix = "Error: Robot with ID"
	if !strings.Contains(output, expectedOutputPrefix) {
		t.Errorf("Expected output to contain '%s', but got:\n%s", expectedOutputPrefix, output)
	}
}

// TestViewCommands tests the "view" and "stop_view" commands.
func TestViewCommands(t *testing.T) {
	setupTest()
	defer setupTest()

	// Test "view" command
	RootCmd.SetArgs([]string{"view"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("view command failed: %v", err)
	}

	if !viewIsRunning {
		t.Fatalf("Expected viewIsRunning to be true after starting view, but it's false")
	}

	// Wait a moment for the goroutine to start.
	time.Sleep(simulationTick + 10*time.Millisecond)

	// Test "stop_view" command
	RootCmd.SetArgs([]string{"stop_view"})
	if err := RootCmd.Execute(); err != nil {
		t.Fatalf("stop_view command failed: %v", err)
	}

	if viewIsRunning {
		t.Fatalf("Expected viewIsRunning to be false after stopping view, but it's true")
	}

	// Verify the 'done' channel is closed by checking if a receive is non-blocking.
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		select {
		case <-done:
			// This is the expected path, the channel is closed.
		case <-time.After(100 * time.Millisecond):
			t.Errorf("Expected done channel to be closed, but it timed out.")
		}
	}()
	wg.Wait()
}
