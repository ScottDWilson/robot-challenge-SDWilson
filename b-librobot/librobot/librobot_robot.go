package librobot

import (
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Robot simulation implementation

// robotImpl implements the Robot interfac for the simulated environment
type robotImpl struct {
	id             string                   // Unique identifier for a robot
	warehouse      *warehouseImpl           // Warehouse robot
	state          RobotState               // Store the current state of the robot; x, y, crate
	canPickCrates  bool                     // Only robots in CrateWarehouses can pick crates
	taskQueue      chan *robotTask          // Channel to queue robot tasks
	cancelChannels map[string]chan struct{} // Map to store cancellation channels for each task
	mu             *sync.Mutex              // Mutex to protect robot's internal state
	stopWorker     chan struct{}            // Channel to signal the worker goroutine to stop
	workerStarted  bool
	isDiagonal     bool // Flag for diagonal movement of robot
}

// robotTask represents an individual task for the robot.
type robotTask struct {
	id         string          // Unique ID for task
	commands   string          // List of commands recieved
	positionCh chan RobotState // Channel to send periodic position updates
	errorCh    chan error      // Channel to send task-specific errors
	cancelCh   chan struct{}   // Channel specific to this task for cancellation
}

// EnqueueTask adds a new task to the robot's queue.
// The tasks will be executed on the robots clock cycle in FIFO queue.
// It returns the task ID and two channels for monitoring: one for position updates and one for errors.
func (r *robotImpl) EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) {

	taskID = uuid.New().String()
	posChan := make(chan RobotState) // Unbuffered, sends immediately
	errChan := make(chan error, 1)   // Buffered, allows error to be sent even if no one is listening immediately

	task := &robotTask{
		id:         taskID,
		commands:   commands,
		positionCh: posChan,
		errorCh:    errChan,
		cancelCh:   make(chan struct{}), // Unbuffered cancellation channel
	}
	r.mu.Lock()

	r.cancelChannels[taskID] = task.cancelCh
	r.taskQueue <- task // Send task to the robot's queue

	r.mu.Unlock() // Unlock after changes to robot

	return taskID, posChan, errChan
}

// CancelTask cancels a task by ID currently enqueued or in progress.
func (r *robotImpl) CancelTask(taskID string) error {

	r.mu.Lock()
	defer r.mu.Unlock()

	cancelCh, ok := r.cancelChannels[taskID]
	if !ok {
		return errors.New("error: Could not cancel task: task not found")
	}

	// Close the channel to signal cancellation. Non-blocking if already closed.
	select {
	case <-cancelCh: // Check if already closed
		// Already closed, do nothing
	default:
		close(cancelCh) // Close the channel
	}

	// Remove from the map regardless, as it's either cancelled or will be shortly.
	delete(r.cancelChannels, taskID)
	return nil
}

// CurrentState returns the current state of the robot.
func (r *robotImpl) CurrentState() RobotState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}

// startWorker starts the robot's dedicated goroutine for processing tasks.
// This should be called only once when the robot is added to the warehouse.
func (r *robotImpl) startWorker() {
	r.mu.Lock()
	if r.workerStarted {
		r.mu.Unlock()
		return // Worker already running
	}
	r.workerStarted = true
	r.mu.Unlock()

	log.Printf("Robot %s worker started at (%d, %d)", r.id, r.state.X, r.state.Y)
	for {
		select {
		case task := <-r.taskQueue:
			r.executeTask(task)
			// Clean up the task's cancel channel after execution/cancellation
			r.mu.Lock()
			delete(r.cancelChannels, task.id)
			r.mu.Unlock()
		case <-r.stopWorker:
			log.Printf("Robot %s worker stopping.", r.id)
			return
		}
	}
}

// executeTask processes a single robotTask.
func (r *robotImpl) executeTask(task *robotTask) {
	log.Printf("Robot %s: Starting task %s with commands: \"%s\"", r.id, task.id, task.commands)
	defer close(task.positionCh) // Close position channel when task is done or aborted
	defer close(task.errorCh)    // Close error channel when task is done or aborted

	commands := parseCommands(task.commands)

	// For diagonal operation, check this command and the next command
	if r.isDiagonal {
		commands = processCommands(commands)
	}

	for i, cmd := range commands {
		select {
		case <-task.cancelCh:
			log.Printf("Robot %s: Task %s cancelled externally after %d commands.", r.id, task.id, i)
			// Send a specific cancellation error if needed, or just let channels close
			select {
			case task.errorCh <- errors.New("task cancelled"):
			default:
				// Error channel might not be listened to, or already closed by external cancel.
			}
			return // Abort task
		default:
			// Continue execution
		}

		err := r.executeCommand(cmd)
		if err != nil {
			log.Printf("Robot %s: Task %s aborted due to error after command '%c': %v", r.id, task.id, cmd, err)
			select {
			case task.errorCh <- err:
			default:
				// Error channel might not be listened to.
			}
			return // Abort task
		}

		// Send current state after successful command
		select {
		case task.positionCh <- r.CurrentState():
		default:
			// If no one is listening, don't block. This can happen if the client
			// stops listening to position updates.
		}

		// Simulate real-time execution
		time.Sleep(CommandExecutionTime)
	}
	log.Printf("Robot %s: Task %s completed successfully.", r.id, task.id)
}

// executeCommand attempts to execute a single robot command.
// It handles movement, boundary checks, and collision detection.
func (r *robotImpl) executeCommand(cmd rune) error {
	r.warehouse.mu.Lock() // Global warehouse lock for grid manipulation
	defer r.warehouse.mu.Unlock()

	r.mu.Lock() // Robot's internal state lock
	defer r.mu.Unlock()

	currentX, currentY := r.state.X, r.state.Y
	newX, newY := currentX, currentY

	// Add new commands here
	switch cmd {
	case 'N':
		newY++
	case 'S':
		newY--
	case 'E':
		newX++
	case 'W':
		newX--
	// Crate interactions
	case 'G':
		if !r.canPickCrates {
			return ErrInvalidWarehouseType //errors.New("error; robot can not pick crates in non-crate warehouse")
		}
		// Get cratewarehouse implementation
		if err := r.grabCrate(); err != nil {
			return err
		}
		log.Printf("Robot %s: Grabbed crate at (%d, %d)", r.id, r.state.X, r.state.Y)

	case 'D':
		if !r.canPickCrates {
			return ErrInvalidWarehouseType //errors.New("error; robot can not pick crates in non-crate warehouse")
		}
		if err := r.dropCrate(); err != nil {
			return err
		}
		log.Printf("Robot %s: Dropped crate at (%d, %d)", r.id, r.state.X, r.state.Y)

	// Phase 3 diagonal motion
	case MoveNorthEast: // Use the defined constant
		newY++
		newX++
	case MoveNorthWest:
		newY++
		newX--
	case MoveSouthEast:
		newY--
		newX++
	case MoveSouthWest:
		newY--
		newX--

	default:
		return fmt.Errorf("unknown command: %c", cmd)
	}

	// Boundary check; uint < 0 is always false
	if newX > (GridSize-1) || newY > (GridSize-1) || newX < 0 || newY < 0 { // newX < 0 || newY < 0 is technically redundant due to uint, but good for clarity if types change
		return ErrOutOfBounds //errors.New("error: command would cause robot to move out of bounds")
	}

	// Collision detection
	if r.warehouse.gridyx[newY][newX] != "" && r.warehouse.gridyx[newY][newX] != r.id {
		// Target cell is occupied by another robot
		return ErrPositionOccupied
	}

	// Update grid: vacate old position, occupy new position
	r.warehouse.gridyx[currentY][currentX] = ""
	r.warehouse.gridyx[newY][newX] = r.id

	// Update robot's internal state
	r.state.X = newX
	r.state.Y = newY

	log.Printf("Robot %s: Moved to (%d, %d)", r.id, r.state.X, r.state.Y)
	return nil
}

// grabCrate Picks crate at current robot position; sets RobotState.HasCrate flag
func (r *robotImpl) grabCrate() error {
	// Check robot carrying crate
	if r.state.HasCrate {
		return ErrRobotHasCrate
	}
	// Check crate exists at position
	if !r.warehouse.cratesyx[r.state.Y][r.state.X] {
		return ErrCrateNotFound
	}
	r.warehouse.cratesyx[r.state.Y][r.state.X] = false
	r.state.HasCrate = true
	return nil
}

// dropCrate Drops crate at current robot position; sets RobotState.HasCrate flag
func (r *robotImpl) dropCrate() error {
	// Check robot carrying crate
	if !r.state.HasCrate {
		return ErrRobotNotCrate
	}
	// Check crate exists at position
	if r.warehouse.cratesyx[r.state.Y][r.state.X] {
		return ErrCrateExists
	}
	r.warehouse.cratesyx[r.state.Y][r.state.X] = true
	r.state.HasCrate = false
	return nil
}

// parseCommands splits a command string into individual command runes. Spilt on whitespace
func parseCommands(commands string) []rune {
	// TODO error handling; if commands are not valid
	var cmds []rune
	for _, r := range commands {
		if r != ' ' { // Ignore whitespace
			cmds = append(cmds, r)
		}
	}
	return cmds
}

// processCommands takes a string of parsed commands and returns a modified string utilising diagonal motion
func processCommands(commands []rune) []rune {
	var processedCmds []rune
	var lastCmd rune

	log.Print("Processing commands to diagonal")

	// We'll use a for loop to iterate through the input commands.
	for _, cmd := range commands {
		// If the last command was a cardinal direction and the current command is orthogonal,
		// we can combine them.
		isCardinal := func(r rune) bool {
			return r == 'N' || r == 'S' || r == 'E' || r == 'W'
		}

		isOrthogonal := func(r1, r2 rune) bool {
			// Check if one is N/S and the other is E/W
			isVertical := r1 == 'N' || r1 == 'S'
			isHorizontal := r2 == 'E' || r2 == 'W'
			return isVertical == isHorizontal // They belong in different groups
		}

		if lastCmd != 0 && isCardinal(lastCmd) && isCardinal(cmd) && isOrthogonal(lastCmd, cmd) {
			// Found a diagonal pair! Combine them into a single rune.
			switch {
			case (lastCmd == 'N' && cmd == 'E') || (lastCmd == 'E' && cmd == 'N'):
				processedCmds = append(processedCmds, MoveNorthEast) // North-East
				lastCmd = 0                                          // Reset the state
			case (lastCmd == 'N' && cmd == 'W') || (lastCmd == 'W' && cmd == 'N'):
				processedCmds = append(processedCmds, MoveNorthWest) // North-West
				lastCmd = 0                                          // Reset the state
			case (lastCmd == 'S' && cmd == 'E') || (lastCmd == 'E' && cmd == 'S'):
				processedCmds = append(processedCmds, MoveSouthEast) // South-East
				lastCmd = 0                                          // Reset the state
			case (lastCmd == 'S' && cmd == 'W') || (lastCmd == 'W' && cmd == 'S'):
				processedCmds = append(processedCmds, MoveSouthWest) // South-West
				lastCmd = 0                                          // Reset the state
			}
		} else {
			// No diagonal pair, so add the last command to the sequence.
			if lastCmd != 0 {
				processedCmds = append(processedCmds, lastCmd)
			}
			lastCmd = cmd
		}
	}

	// Append any remaining command
	if lastCmd != 0 {
		processedCmds = append(processedCmds, lastCmd)
	}

	return processedCmds
}
