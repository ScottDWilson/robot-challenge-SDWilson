package librobot

import (
	"errors"
	"sync"
)

// Robot simulation implementation

// robotImpl implements the Robot interfac for the simulated environment
type robotImpl struct {
	id            string         // Unique identifier for a robot
	warehouse     *warehouseImpl // Warehouse robot
	state         RobotState     // Store the current state of the robot; x, y, crate
	mu            *sync.Mutex    // Mutex to protect robot's internal state
	stopWorker    chan struct{}  // Channel to signal the worker goroutine to stop
	workerStarted bool
}

// EnqueueTask adds a new task to the robot's queue.
// The tasks will be executed on the robots clock cycle in FIFO queue.
// It returns the task ID and two channels for monitoring: one for position updates and one for errors.
func (r *robotImpl) EnqueueTask(commands string) (taskID string, position chan RobotState, err chan error) {
	temp := make(chan RobotState)
	temp_err_chan := make(chan error, 1)
	temp_err_chan <- errors.New("Enqueue Task not implemented yet")
	return "NA", temp, temp_err_chan
}

// CancelTask cancels a task by ID currently enqueued or in progress.
func (r *robotImpl) CancelTask(taskID string) error {
	return errors.New("Cancel task not implemented yet")
}

// CurrentState returns the current state of the robot.
func (r *robotImpl) CurrentState() RobotState {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.state
}
