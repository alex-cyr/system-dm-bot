# System DM Bot: A&R Agent

automation engine for System Films' localized outreach pipeline. 

This repository contains the Golang architecture for a sovereign **Vision-Language-Action (VLA) Agent**. Unlike traditional bots, system-dm-bot is designed to operate seamlessly within highly guarded social platforms without triggering automated anti-bot security systems.

## Why

Historically, web automation relied on reading a website's background code (DOM manipulation) or reverse-engineering hidden APIs. In the current digital landscape, trillion-dollar AI systems employed by social networks instantly detect and ban headless browsers and scraper scripts. 

It does not read code; it reads pixels. It does not send HTTP requests to internal endpoints; it physically commands the operating system's hardware mouse and keyboard drivers. 
1. **The Eyes:** Vertex AI (Gemini) takes a screenshot of the virtual monitor and returns the exact [X, Y] pixel coordinates of the interface elements we want to interact with.
2. **The Hands:** We utilize low-level C-bindings (`RobotGo`) to emulate human kinematics—moving the mouse along randomized Bezier curves and typing with millisecond, human-like cadences.

To the platform's security algorithms, system-dm-bot is indistinguishable from a System Films A&R representative sitting at a computer in Atlanta, reading a screen, and physically clicking a mouse.

---

## How to Open & Run the Project

This is a strictly Golang and Google Cloud Platform (GCP) stack. To run this codebase locally, you must follow these prerequisites carefully.

### 1. Prerequisites
- **Golang (1.21+)**: Ensure Go is installed on your machine.
- **A C-Compiler**: Because `RobotGo` uses CGO to communicate with your operating system's hardware drivers, **you must have a C-Compiler installed**. 
  - *On Windows:* Install [MinGW-w64](https://www.mingw-w64.org/) or TDM-GCC, and ensure it is added to your System PATH. If you try to run `go build` without this, you will get errors like `undefined: Bitmap`.
  - *On Linux/Mac:* Install `gcc`.
- **Google Cloud Auth**: You need your GCP credentials configured locally to test the Vertex AI vision prompts.

### 2. Project Structure

```text
system-dm-bot/
├── cmd/
│   ├── s66-bot/          # The main Infinite State Machine (FSM) loop. The brain.
│   ├── test-hardware/    # Safe, isolated test script for mouse/keyboard emulation.
│   └── test-vision/      # Safe, isolated test script for Vertex AI spatial coordinate mapping.
├── pkg/
│   ├── cognition/        # (WIP) SMOD/RAG Memory injection for crafting the perfect A&R pitch.
│   ├── hardware/         # RobotGo bindings for physical actuation (MoveSmooth, TypeStrDelay).
│   ├── optics/           # Vertex AI SDK client to process screen bitmaps.
│   └── pipeline/         # (WIP) Firestore logging to prevent duplicate messaging.
```

### 3. Testing the Build
Before integrating everything into the main loop, we test the "Hands" and "Eyes" independently.

**Test the Hands:**

*Note: Because RobotGo requires C-bindings, you must enable CGO and use the correct build mode to prevent Windows executable corruption.*

```bash
# 1. Enable CGO locally
go env -w CGO_ENABLED=1

# 2. Build the executable safely (bypassing PE/COFF linker bugs)
go build -buildmode=exe -ldflags="-s -w" -o test_hands.exe ./cmd/test-hardware/main.go

# 3. Run the compiled test
.\test_hands.exe
```
*(Warning: The moment you press enter on step 3, take your hand off the mouse. It will physically move your cursor and type on your screen).*

**Test the Eyes:**
```bash
go run ./cmd/test-vision/main.go
```
*(Ensure your GCP environment variables are set so it can authenticate with Vertex AI).*

---

##  Team Help

We have cleanly separated the logic so multiple team members can build this.

### 1. The Prompt Engineers (Focus: `pkg/optics/vision.go`)
We need the Vertex AI spatial prompts to be bulletproof. Your job is to upload screenshots of Instagram DMs to Google AI Studio, figure out the exact text prompt needed to make Gemini return the perfect bounding box for the "New Message" dot or the "Reply" text box, and integrate those prompts here.

### 2. The Cultural Strategists (Focus: `pkg/cognition/smod.go`)
An agent's physical stealth is useless if it sounds like a corporate bot. This module handles the generation of the actual DM reply. We need logic that takes the prospect's profile data and drafts an authentic, culturally tuned pitch for a System Films music video shoot (e.g., pricing, locations, aesthetic matching). 

### 3. The Kinematic Engineers (Focus: `pkg/hardware/motor.go`)
Tune the RobotGo parameters. Adjust the Bezier curve mathematical bounds, tweak the typing delay randomization, and ensure the physical movements look as humanly imperfect as possible to guarantee we never get banned.

### 4. The Database Architects (Focus: `pkg/pipeline/memory.go`)
We need to connect this to Google Cloud Firestore. Every time a DM is sent, it must be logged so the bot never double-pitches the same artist twice. 
