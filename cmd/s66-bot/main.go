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

// Agent holds the FSM state, tools, and memory parameters.
type Agent struct {
	Ctx         context.Context
	Vision      *optics.VisionClient
	ScrollCount int
}

// State defines the function signature for our FSM.
type State func(a *Agent) (State, error)

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

	agent := &Agent{
		Ctx:         ctx,
		Vision:      vc,
		ScrollCount: 0,
	}

	var currentState State = StateObserve

	// The infinite FSM loop
	for currentState != nil {
		nextState, err := currentState(agent)
		if err != nil {
			fmt.Printf("Error in state machine: %v\n", err)
			break
		}
		currentState = nextState
	}
	
	fmt.Println("S66 Reachout Bot Terminated")
}

// StateObserve captures the screen and checks for the unread blue dot.
func StateObserve(a *Agent) (State, error) {
	fmt.Println("\n[STATE] Observe: Scanning for unread messages...")
	
	// Give the UI a moment to settle before taking screenshot
	time.Sleep(2 * time.Second)

	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	prompt := "Find the unread message blue dot indicator. If it exists, return only the bounding box [ymin, xmin, ymax, xmax]. If there are no unread messages, return NONE."
	coords, err := a.Vision.LocateElement(a.Ctx, imgBytes, prompt)
	if err != nil {
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
	absoluteX := int((xCenter / 1000.0) * 1920.0)
	absoluteY := int((yCenter / 1000.0) * 1080.0)

	fmt.Printf("Found unread message at X: %d, Y: %d\n", absoluteX, absoluteY)
	
	// Move and click!
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()

	// Wait 6 seconds for internet lag to safely load the chat page
	time.Sleep(6 * time.Second)

	// Reset scroll memory since we found a lead
	a.ScrollCount = 0

	return StateReadAndReply, nil
}

// StateScroll scrolls the mouse down to find more DMs.
func StateScroll(a *Agent) (State, error) {
	fmt.Printf("\n[STATE] Scroll: Moving down the DM list... (Scroll Count: %d/3)\n", a.ScrollCount)
	
	if a.ScrollCount >= 3 {
		fmt.Println("Hit max scroll depth. Refreshing page to catch new DMs.")
		hardware.RefreshPage()
		a.ScrollCount = 0
		return StateObserve, nil
	}

	hardware.ScrollDown()
	a.ScrollCount++
	
	// Let the screen settle
	time.Sleep(2 * time.Second)
	return StateObserve, nil
}

// StateReadAndReply reads the conversation, filters for leads, and drafts a reply.
func StateReadAndReply(a *Agent) (State, error) {
	fmt.Println("\n[STATE] Read & Reply: Analyzing conversation...")
	
	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	// Cognitive Filter: YES / NO
	filterPrompt := "Read this Instagram conversation. Is this user a potential new lead for a music video? If they are a friend, a past client, or not a lead, reply with exactly 'NO'. If they are a new potential lead, reply with exactly 'YES'."
	decision, err := a.Vision.AnalyzeImage(a.Ctx, imgBytes, filterPrompt)
	if err != nil {
		fmt.Println("Error analyzing image, backing out safely.")
		return StateReset, nil
	}

	if strings.Contains(strings.ToUpper(decision), "NO") {
		fmt.Println("Cognitive Filter Decision: NO. This is not a lead. Backing out.")
		return StateReset, nil
	}

	fmt.Println("Cognitive Filter Decision: YES! Proceeding with pitch.")

	// Locate the input box
	prompt := "Find the 'Message...' text input box at the bottom of the chat. Return only the bounding box [ymin, xmin, ymax, xmax]."
	coords, err := a.Vision.LocateElement(a.Ctx, imgBytes, prompt)
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
func StateReset(a *Agent) (State, error) {
	fmt.Println("\n[STATE] Reset: Returning to inbox...")
	
	imgBytes, err := hardware.CaptureScreen()
	if err != nil {
		return nil, err
	}

	prompt := "Find the 'Back' arrow or button in the top left. Return only the bounding box."
	coords, err := a.Vision.LocateElement(a.Ctx, imgBytes, prompt)
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
