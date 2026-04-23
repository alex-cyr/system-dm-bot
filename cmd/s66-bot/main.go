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
	SkillPrompt string
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

	skillBytes, err := os.ReadFile(".agents/skills/instagram-outreach/SKILL.md")
	if err != nil {
		fmt.Println("Warning: Could not load SKILL.md, proceeding without dynamic personality.")
	}

	agent := &Agent{
		Ctx:         ctx,
		Vision:      vc,
		ScrollCount: 0,
		SkillPrompt: string(skillBytes),
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
	
	// Park the mouse in the corner to prevent CSS hover menus from expanding and blocking the screenshot
	hardware.ParkMouse()

	// Give the UI a moment to settle before taking screenshot
	time.Sleep(2 * time.Second)

	screenWidth, screenHeight := hardware.GetScreenDimensions()
	
	// Crop bounds for Inbox Pane: 4% to 37% of screen width. Height: 15% down to bottom.
	// We use 33% width to ensure the blue dots on the far right are captured even if the browser is zoomed to 150%.
	inboxX := int(float64(screenWidth) * 0.04)
	inboxW := int(float64(screenWidth) * 0.33)
	inboxY := int(float64(screenHeight) * 0.15)
	inboxH := int(float64(screenHeight) * 0.85)

	imgBytes, err := hardware.CaptureRect(inboxX, inboxY, inboxW, inboxH)
	if err != nil {
		return nil, err
	}

	prompt := "Find an unread incoming direct message row in this inbox list. Unread messages have a blue dot on the far right, the text is brighter, and they do NOT start with 'You:'. Return the bounding box [ymin, xmin, ymax, xmax] of the ENTIRE row for the unread message. If there are no unread messages, return NONE."
	coords, err := a.Vision.LocateElement(a.Ctx, imgBytes, prompt)
	if err != nil {
		if strings.Contains(err.Error(), "could not find bounding box") || strings.Contains(err.Error(), "NONE") {
			fmt.Println("No unread messages found in this viewport.")
			return StateScroll, nil
		}
		return nil, err
	}

	// Calculate pixel center RELATIVE to the crop (only care about Y now)
	yCenter := (coords[0] + coords[2]) / 2
	
	// Convert normalized (0.0 - 1000.0) to absolute pixels relative to crop
	cropY := int((yCenter / 1000.0) * float64(inboxH))

	// Translate crop-relative Y to true global screen Y
	absoluteY := inboxY + cropY

	// Hardening Phase 5: Completely ignore Vertex AI's X coordinate to prevent horizontal hallucinations.
	// Hardcode the X click to the dead center of the Inbox Pane.
	absoluteX := inboxX + (inboxW / 2)

	fmt.Printf("Found unread message at X: %d, Y: %d\n", absoluteX, absoluteY)
	
	// Move and click!
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()

	// Wait 3 seconds for internet lag to safely load the chat page (reduced from 6s)
	time.Sleep(3 * time.Second)

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

	// Phase 7: Force mouse to hover over the exact safe zone of the inbox pane before scrolling
	hardware.ParkMouse()
	time.Sleep(500 * time.Millisecond)

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
	filterPrompt := "Read this Instagram conversation. Is this user a potential new lead for a music video? Pay extremely close attention to Atlanta slang and informal DMs. Examples of leads: 'whats da tixket' (what's the price), 'how much for a vid', 'yall shooting?', 'wya', 'send rates', 'need sum work done'. Any inquiry about price, info, availability, or video packages is a lead. If they are a friend, a past client, or explicitly not interested, reply with exactly 'NO'. If they are a new potential lead, reply with exactly 'YES'."
	decision, err := a.Vision.AnalyzeImage(a.Ctx, imgBytes, filterPrompt)
	if err != nil {
		fmt.Println("Error analyzing image, backing out safely.")
		return StateObserve, nil
	}

	if strings.Contains(strings.ToUpper(decision), "NO") {
		fmt.Println("Cognitive Filter Decision: NO. This is not a lead. Backing out.")
		return StateObserve, nil
	}

	fmt.Println("Cognitive Filter Decision: YES! Proceeding with pitch.")

	// Locate the input box using highly precise text targeting
	screenWidth, screenHeight := hardware.GetScreenDimensions()
	
	// Crop bounds for Message Input Area: Bottom 20% of the right 70% of the screen
	chatX := int(float64(screenWidth) * 0.30)
	chatW := int(float64(screenWidth) * 0.70)
	chatY := int(float64(screenHeight) * 0.80)
	chatH := int(float64(screenHeight) * 0.20)

	msgBytes, err := hardware.CaptureRect(chatX, chatY, chatW, chatH)
	if err != nil {
		return nil, err
	}

	prompt := "Find the exact word 'Message...' inside the text input area. Return only the bounding box."
	coords, err := a.Vision.LocateElement(a.Ctx, msgBytes, prompt)
	if err != nil {
		fmt.Println("Could not find the message input box. Backing out.")
		return StateObserve, nil
	}

	yCenter := (coords[0] + coords[2]) / 2
	
	cropY := int((yCenter / 1000.0) * float64(chatH))
	
	// Hardening Phase 5: Ignore Vertex AI's X coordinate and hardcode to the center of the Chat Pane
	absoluteX := chatX + (chatW / 2)
	absoluteY := chatY + cropY

	// Click into the message box
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()
	time.Sleep(200 * time.Millisecond)
	hardware.Click() // Phase 6: Double-click to absolutely guarantee focus lock
	
	time.Sleep(1 * time.Second)

	// Type the dynamically generated pitch
	fmt.Println("Cognitive Engine: Drafting personalized pitch...")
	draftPrompt := fmt.Sprintf("Based on the following personality rules, draft a 1 to 2 sentence opening pitch to this Instagram user. Output ONLY the raw pitch text. Do not include quotes, markdown, or any other text.\n\nRules:\n%s", a.SkillPrompt)
	
	pitchText, err := a.Vision.AnalyzeImage(a.Ctx, imgBytes, draftPrompt)
	if err != nil || pitchText == "" {
		fmt.Println("Failed to draft pitch dynamically. Using fallback.")
		pitchText = "Hey! We are an Atlanta video crew shooting $400 cinema-grade music videos. Let me know if you are interested in a shoot!"
	}

	// Clean up any stray quotes or whitespace from the LLM
	pitchText = strings.TrimSpace(pitchText)
	pitchText = strings.Trim(pitchText, "\"")
	pitchText = strings.Trim(pitchText, "'")

	err = hardware.PasteText(pitchText)
	if err != nil {
		fmt.Printf("Warning: Failed to paste text via clipboard: %v\n", err)
		// Fallback to typing if clipboard fails for some reason
		hardware.TypeStrDelay(pitchText)
	}
	time.Sleep(1 * time.Second)
	
	// Hit Enter to send
	hardware.PressEnter()
	time.Sleep(2 * time.Second) // wait for send

	// Desktop Split-Pane UI doesn't have a back button. Return straight to observe.
	return StateObserve, nil
}
