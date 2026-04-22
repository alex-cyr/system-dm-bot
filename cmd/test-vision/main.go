package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alex-cyr/system-dm-bot/pkg/hardware"
	"github.com/alex-cyr/system-dm-bot/pkg/optics"
)

func main() {
	fmt.Println("Starting Local Optics Test...")
	ctx := context.Background()

	// 1. Capture screen with RobotGo
	fmt.Println("Using RobotGo to capture a test screenshot...")
	imageBytes, err := hardware.CaptureScreen()
	if err != nil {
		fmt.Printf("Failed to capture screen: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Captured image of size: %d bytes\n", len(imageBytes))

	// 2. Initialize Vertex Vision Client
	// Note: You must set up GCP credentials to actually run a live query
	client, err := optics.NewVisionClient("system-dm-bot", "us-central1")
	if err != nil {
		fmt.Printf("Failed to initialize Vision Client: %v\n", err)
		os.Exit(1)
	}

	// 3. Locate element via mock
	prompt := "Find the application icon. Output bounding box."
	coords, err := client.LocateElement(ctx, imageBytes, prompt)
	if err != nil {
		fmt.Printf("Failed to locate element: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Vertex AI returned normalized coordinates: %v\n", coords)
	
	// Convert coordinates to absolute (assuming 1920x1080 screen for test)
	absX := (coords[1] / 1000.0) * 1920.0
	absY := (coords[0] / 1000.0) * 1080.0
	fmt.Printf("Calculated Absolute Pixels: X: %f, Y: %f\n", absX, absY)
	fmt.Println("Local Optics Test Completed.")
}
