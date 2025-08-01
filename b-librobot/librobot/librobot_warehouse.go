package librobot

import (
	"log"
	"sync"
)

// An implementation of the warehouse for simulation purposes
// Provides options to override interface with alternative implementations later

// NewWarehouse creates and returns a new simulated Warehouse instance.
// The warehouse grid dimensions are defined by GridSize.
func NewWarehouse() Warehouse {
	w := &warehouseImpl{
		robots: make(map[string]*robotImpl),
		gridyx: [GridSize + 1][GridSize + 1]string{}, // Initialize with empty strings
		mu:     &sync.RWMutex{},                      // Controls access to changing settings so only one at a time
	}
	log.Println("New Warehouse created.")
	return w
}

// warehouseImpl implements the Warehouse and later CrateWarehouse interfaces.
type warehouseImpl struct {
	// A map of the robot ID to the robot implementation
	robots map[string]*robotImpl
	// Gridyx stores the ID of the robot occupying a cell, or an empty string if vacant.
	// gridyx[y][x] for easier access: grid[row][column]
	gridyx [GridSize + 1][GridSize + 1]string
	mu     *sync.RWMutex // Mutex to protect access to robots and grid
	// crates stores whether a crate is present at a specific "x,y" coordinate string.
	crates map[string]bool // Only for CrateWarehouse
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
func AddRobot(w Warehouse, initialX, initialY uint) (Robot, error) {
	// TODO; add robot implementation to list, check for errors

}
