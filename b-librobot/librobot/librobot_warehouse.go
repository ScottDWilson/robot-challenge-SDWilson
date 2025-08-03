package librobot

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/google/uuid" // Create unique identifier for each warehouse, robot
)

// An implementation of the warehouse for simulation purposes
// Provides options to override interface with alternative implementations later

// NewWarehouse creates and returns a new simulated Warehouse instance.
// The warehouse grid dimensions are defined by GridSize.
func NewWarehouse() Warehouse {
	w := &warehouseImpl{
		robots:     make(map[string]*robotImpl),
		gridyx:     [GridSize + 1][GridSize + 1]string{}, // Initialize with empty strings
		mu:         &sync.RWMutex{},                      // Controls access to changing settings so only one at a time
		has_crates: false,
	}
	log.Println("New Warehouse created.")
	return w
}

// NewCrateWarehouse creates a new warehouse with no robots. This warehouse can accept crates
// This function now returns the new CrateWarehouse interface.
func NewCrateWarehouse() CrateWarehouse {
	cw := &warehouseImpl{
		robots: make(map[string]*robotImpl),
		gridyx: [GridSize + 1][GridSize + 1]string{}, // Initialize with empty strings
		mu:     &sync.RWMutex{},                      // Controls access to changing settings so only one at a time
		// cratesyx defaults to false
		has_crates: true,
	}
	log.Println("New Crate Warehouse created.")
	return cw
}

// warehouseImpl implements the Warehouse and later CrateWarehouse interfaces.
type warehouseImpl struct {
	// A map of the robot ID to the robot implementation
	robots map[string]*robotImpl
	// Gridyx stores the ID of the robot occupying a cell, or an empty string if vacant.
	// gridyx[y][x] for easier access: grid[row][column]
	gridyx     [GridSize + 1][GridSize + 1]string
	mu         *sync.RWMutex                    // Mutex to protect access to robots and grid
	cratesyx   [GridSize + 1][GridSize + 1]bool // 2D array of crate locations. Refactor if warehouse can be huge for memory optimisation
	has_crates bool
}

// Robots returns a list of all robots currently in the warehouse.
func (w *warehouseImpl) Robots() []Robot {
	w.mu.RLock()
	defer w.mu.RUnlock()

	robotList := make([]Robot, 0, len(w.robots))
	for _, r := range w.robots {
		robotList = append(robotList, r)
	}
	return robotList
}

// AddRobot adds a new robot to the warehouse at the specified initial coordinates.
// It returns the new Robot instance and an error if the position is invalid or occupied.
func AddRobot(w Warehouse, initialX, initialY uint, namedID string) (Robot, error) {
	// Type assertion to get the concrete warehouseImpl
	wh, ok := w.(*warehouseImpl)
	if !ok {
		return nil, errors.New("invalid warehouse type")
	}

	isCrateWarehouse := wh.has_crates

	// Safe access
	wh.mu.Lock()
	defer wh.mu.Unlock()

	// Check desired initial position is within the specified grid size (10x10 default)
	if initialX > GridSize || initialY > GridSize {
		return nil, errors.New("error: initial X and Y are out of bounds")
	}
	if wh.gridyx[initialY][initialX] != "" {
		return nil, errors.New("error: a robot exists at this positin")
	}

	robotID := ""

	// Use named ID if given; if not, use UUID
	if namedID == "" {
		// Initialise robot Uuid
		robotID = uuid.New().String()
	} else {
		robotID = namedID
	}
	// Create robot with defaults
	robot := &robotImpl{
		id:             robotID,
		warehouse:      wh,
		state:          RobotState{X: initialX, Y: initialY, HasCrate: false},
		canPickCrates:  isCrateWarehouse,
		taskQueue:      make(chan *robotTask, 100),     // Buffered channel for tasks
		cancelChannels: make(map[string]chan struct{}), // Initialise
		mu:             &sync.Mutex{},
		stopWorker:     make(chan struct{}),
	}

	// Add robot to list robots in this warehouse
	wh.robots[robotID] = robot
	// Add robot to grid
	wh.gridyx[initialY][initialX] = robotID

	// Start worker
	go robot.startWorker()

	return robot, nil
}

// AddDiagonalRobot adds a new robot to the warehouse at the specified initial coordinates.
// This robot has the capability to move diagonally in the grid when coordinates are in the correct sequence.
// It returns the new Robot instance and an error if the position is invalid or occupied.
func AddDiagonalRobot(w Warehouse, initialX, initialY uint, namedID string) (Robot, error) {
	// Type assertion to get the concrete warehouseImpl
	wh, ok := w.(*warehouseImpl)
	if !ok {
		return nil, ErrInvalidWarehouseType
	}

	isCrateWarehouse := wh.has_crates

	// Safe access
	wh.mu.Lock()
	defer wh.mu.Unlock()

	// Check desired initial position is within the specified grid size (10x10 default)
	if initialX > GridSize || initialY > GridSize {
		return nil, errors.New("error: initial X and Y are out of bounds")
	}
	if wh.gridyx[initialY][initialX] != "" {
		return nil, errors.New("error: a robot exists at this positin")
	}
	robotID := ""
	// Use named ID if given; if not, use UUID
	if namedID == "" {
		// Initialise robot Uuid
		robotID = uuid.New().String()
	} else {
		robotID = namedID
	}

	// Create robot with defaults
	robot := &robotImpl{
		id:             robotID,
		warehouse:      wh,
		state:          RobotState{X: initialX, Y: initialY, HasCrate: false},
		canPickCrates:  isCrateWarehouse,
		taskQueue:      make(chan *robotTask, 100),     // Buffered channel for tasks
		cancelChannels: make(map[string]chan struct{}), // Initialise
		mu:             &sync.Mutex{},
		stopWorker:     make(chan struct{}),
		isDiagonal:     true,
	}

	// Add robot to list robots in this warehouse
	wh.robots[robotID] = robot
	// Add robot to grid
	wh.gridyx[initialY][initialX] = robotID

	// Start worker
	go robot.startWorker()

	return robot, nil
}

// GetRobot retrieves a robot by its ID.
func (w *warehouseImpl) GetRobot(id string) (Robot, bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	r, ok := w.robots[id]
	return r, ok
}

// Adds a crate to the specified x y coordinates
func (cw *warehouseImpl) AddCrate(x uint, y uint) error {
	if !cw.has_crates {
		return ErrInvalidWarehouseType
	}

	cw.mu.Lock()
	defer cw.mu.Unlock()

	if x > GridSize || y > GridSize {
		return ErrOutOfBounds
	}
	// Check for a crate with a direct array lookup.
	if cw.cratesyx[y][x] {
		return ErrCrateExists
	}
	cw.cratesyx[y][x] = true
	log.Printf("Crate added at (%d, %d).", x, y)
	return nil
}

// Deletes a crate at the specified x y coordinates
func (cw *warehouseImpl) DelCrate(x uint, y uint) error {
	if !cw.has_crates {
		return ErrInvalidWarehouseType
	}

	cw.mu.Lock()
	defer cw.mu.Unlock()

	if x > GridSize || y > GridSize {
		return ErrOutOfBounds
	}
	// Check for a crate with a direct array lookup.
	if !cw.cratesyx[y][x] {
		return ErrCrateNotFound
	}
	cw.cratesyx[y][x] = false
	log.Printf("Crate deleted from (%d, %d).", x, y)
	return nil
}

// ClearScreen uses ANSI escape codes to clear the terminal screen.
func ClearScreen() {
	// \033[H: Moves the cursor to the top-left corner
	// \033[2J: Clears the entire screen
	fmt.Print("\033[H\033[2J")
}

// Render draws the current state of the warehouse. It does not explicity clear screen
func Render(w CrateWarehouse, robot_map map[string]Robot) {
	// Retrieve warehouse implementation
	wh, ok := w.(*warehouseImpl)
	if !ok {
		return
	}

	// Create a 2D array to represent the grid
	grid := make([][]string, GridSize)
	for i := range grid {
		grid[i] = make([]string, GridSize)
		for j := range grid[i] {
			grid[i][j] = " - " // Default empty space

			// Check crate and Add it
			if wh.has_crates {
				if wh.cratesyx[i][j] {
					grid[i][j] = "[C]"
				}
			}
		}
	}

	// Place robots on the grid (overwriting crates if necessary)
	for id, robot := range wh.robots {
		state := robot.CurrentState()
		if state.Y < GridSize && state.X < GridSize {
			label := id
			//symbol := fmt.Sprintf("R%d ", i) // e.g., "R0 "
			symbol := label[0:2]
			if state.HasCrate {
				//symbol = fmt.Sprintf("R%d*", i) // e.g., "R0*"
				symbol += "*"
			} else if wh.cratesyx[state.Y][state.X] {
				symbol += "_"
			}
			grid[state.Y][state.X] = symbol
		}
	}

	// Build the output string and print
	var builder strings.Builder
	builder.WriteString("--- Warehouse Real-Time View ---\n")
	// Display grid, with 0,0 as the bottom left corner for good UX
	for y := int(GridSize - 1); y >= 0; y-- {
		for x := uint(0); x < GridSize; x++ {
			builder.WriteString(grid[y][x])
		}
		builder.WriteString("\n")
	}
	builder.WriteString("--------------------------------\n")
	//builder.WriteString("Enter command >>.\n")
	fmt.Print(builder.String())
}
