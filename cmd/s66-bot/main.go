package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/alex-cyr/system-dm-bot/pkg/hardware"
	"github.com/alex-cyr/system-dm-bot/pkg/optics"
)

// State defines the function signature for our FSM.
type State func(ctx context.Context, vc *optics.VisionClient) (State, error)

func main() {
	fmt.Println("S66 Reachout Bot Initialized (VLA Sovereign Agent)")

	ctx := context.Background()

	// Initialize Motor
	hardware.InitMotor()

	// Initialize Vision Client
	fmt.Println("Connecting to Vertex AI Power Plant...")
	vc, err := optics.NewVisionClient("matrix-esa-production", "us-central1")
	if err != nil {
		fmt.Printf("Failed to init vision: %v\n", err)
		os.Exit(1)
	}

	var currentState State = StateObserve

	// The infinite FSM loop
	for currentState != nil {
		nextState, err := currentState(ctx, vc)
		if err != nil {
			fmt.Printf("Error in state machine: %v\n", err)
			break
		}
		currentState = nextState
	}
	
	fmt.Println("S66 Reachout Bot Terminated")
}

// StateObserve captures the screen and checks for the unread blue dot.
func StateObserve(ctx context.Context, vc *optics.VisionClient) (State, error) {
	fmt.Println("\n[STATE] Observe: Scanning for unread messages...")
	
	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	prompt := "Find the unread message blue dot indicator. If it exists, return only the bounding box [ymin, xmin, ymax, xmax]. If there are no unread messages, return NONE."
	coords, err := vc.LocateElement(ctx, imgBytes, prompt)
	if err != nil {
		// If Gemini failed to find coordinates, it either returned NONE or errored.
		// We'll treat this as "no unread messages found" and scroll down.
		if strings.Contains(err.Error(), "could not find bounding box") || strings.Contains(err.Error(), "NONE") {
			fmt.Println("No unread messages found in this viewport.")
			return StateScroll, nil
		}
		return nil, err
	}

	// Calculate absolute pixel center
	yCenter := (coords[0] + coords[2]) / 2
	xCenter := (coords[1] + coords[3]) / 2
	
	// Convert normalized (0.0 - 1000.0) to absolute screen pixels.
	// Note: LocateElement returns normalized float64 but currently Vertex JSON outputs standard integers 0-1000.
	// We need actual screen bounds, but for this demo, we assume the user's screen is 1920x1080.
	// We'll normalize the 1000-scale coordinates to the 1080p screen.
	absoluteX := int((xCenter / 1000.0) * 1920.0)
	absoluteY := int((yCenter / 1000.0) * 1080.0)

	// Inject the coordinates into the context or just pass them globally (for this simple MVP we will just move instantly)
	fmt.Printf("Found unread message at X: %d, Y: %d\n", absoluteX, absoluteY)
	
	// Move and click!
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()

	// Wait for internet lag to load the chat page
	time.Sleep(4 * time.Second)

	return StateReadAndReply, nil
}

// StateScroll scrolls the mouse down to find more DMs.
func StateScroll(ctx context.Context, vc *optics.VisionClient) (State, error) {
	fmt.Println("\n[STATE] Scroll: Moving down the DM list...")
	hardware.ScrollDown()
	// Let the screen settle
	time.Sleep(2 * time.Second)
	return StateObserve, nil
}

// StateReadAndReply reads the conversation and drafts/sends a reply.
func StateReadAndReply(ctx context.Context, vc *optics.VisionClient) (State, error) {
	fmt.Println("\n[STATE] Read & Reply: Analyzing conversation...")
	
	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	// In a full production bot, we would query the SKILL.md here to parse the conversation and generate a response.
	// For this Phase 3 integration, we use Vertex to find the message input box.
	prompt := "Find the 'Message...' text input box at the bottom of the chat. Return only the bounding box [ymin, xmin, ymax, xmax]."
	coords, err := vc.LocateElement(ctx, imgBytes, prompt)
	if err != nil {
		fmt.Println("Could not find the message input box. Backing out.")
		return StateReset, nil
	}

	yCenter := (coords[0] + coords[2]) / 2
	xCenter := (coords[1] + coords[3]) / 2
	absoluteX := int((xCenter / 1000.0) * 1920.0)
	absoluteY := int((yCenter / 1000.0) * 1080.0)

	// Click into the message box
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()

	// Type the pre-drafted pitch
	hardware.TypeStrDelay("Hey! We are casting for a music video in Atlanta next week. Let me know if you are free!")
	time.Sleep(1 * time.Second)
	
	// Hit Enter to send
	hardware.TypeStrDelay("\n")
	time.Sleep(2 * time.Second) // wait for send

	return StateReset, nil
}

// StateReset clicks back to the main DM board.
func StateReset(ctx context.Context, vc *optics.VisionClient) (State, error) {
	fmt.Println("\n[STATE] Reset: Returning to inbox...")
	
	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	prompt := "Find the 'Back' arrow or button in the top left. Return only the bounding box."
	coords, err := vc.LocateElement(ctx, imgBytes, prompt)
	if err != nil {
		fmt.Println("Could not find the back button. Re-observing...")
		return StateObserve, nil
	}

	yCenter := (coords[0] + coords[2]) / 2
	xCenter := (coords[1] + coords[3]) / 2
	absoluteX := int((xCenter / 1000.0) * 1920.0)
	absoluteY := int((yCenter / 1000.0) * 1080.0)

	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()

	// Wait for the inbox to load
	time.Sleep(3 * time.Second)
	
	return StateObserve, nil
}
