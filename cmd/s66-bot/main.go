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

	// Hardening Phase 4: Offset X to click the profile picture/name instead of the far-right blue dot void
	absoluteX -= 200
	if absoluteX < 10 {
		absoluteX = 10
	}

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
		return StateObserve, nil
	}

	if strings.Contains(strings.ToUpper(decision), "NO") {
		fmt.Println("Cognitive Filter Decision: NO. This is not a lead. Backing out.")
		return StateObserve, nil
	}

	fmt.Println("Cognitive Filter Decision: YES! Proceeding with pitch.")

	// Locate the input box using highly precise text targeting
	prompt := "Find the exact word 'Message...' inside the text input area at the bottom. Return only the bounding box."
	coords, err := a.Vision.LocateElement(a.Ctx, imgBytes, prompt)
	if err != nil {
		fmt.Println("Could not find the message input box. Backing out.")
		return StateObserve, nil
	}

	yCenter := (coords[0] + coords[2]) / 2
	xCenter := (coords[1] + coords[3]) / 2
	absoluteX := int((xCenter / 1000.0) * 1920.0)
	absoluteY := int((yCenter / 1000.0) * 1080.0)

	// Click into the message box
	hardware.MoveSmooth(absoluteX, absoluteY)
	hardware.Click()
	
	// Phase 6: Guarantee the browser registers the focus before the bot starts typing
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

	hardware.TypeStrDelay(pitchText)
	time.Sleep(1 * time.Second)
	
	// Hit Enter to send
	hardware.PressEnter()
	time.Sleep(2 * time.Second) // wait for send

	// Desktop Split-Pane UI doesn't have a back button. Return straight to observe.
	return StateObserve, nil
}
