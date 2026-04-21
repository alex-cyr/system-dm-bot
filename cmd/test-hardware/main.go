package main

import (
	"fmt"
	"time"

	"github.com/alex-cyr/system-dm-bot/pkg/hardware"
)

func main() {
	fmt.Println("Starting Local Hardware Test in 3 seconds...")
	fmt.Println("Please move your cursor to a safe spot. The mouse will move autonomously.")
	time.Sleep(3 * time.Second)

	hardware.InitMotor()

	// 1. Smooth move to a somewhat central location (e.g., 500, 500)
	hardware.MoveSmooth(500, 500)

	// 2. Perform a click
	hardware.Click()

	// 3. Type a simulated message with human delay
	hardware.TypeStrDelay("Testing S66 Reachout Bot.")

	fmt.Println("Local Hardware Test Completed.")
}
