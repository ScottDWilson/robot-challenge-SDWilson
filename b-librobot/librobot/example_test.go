package librobot_test

import (
	"fmt"

	"librobot"
)

func ExampleRobot_basicUsage() {
	warehouse := librobot.NewWarehouse()

	fmt.Printf("Warehouse created %v", warehouse.Robots())
}
