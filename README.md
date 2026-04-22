# System DM Bot: A&R Agent

automation engine for System Films' localized outreach pipeline. 

This repository contains the Golang architecture for a sovereign **Vision-Language-Action (VLA) Agent**. Unlike traditional bots, system-dm-bot is designed to operate seamlessly within highly guarded social platforms without triggering automated anti-bot security systems.

## Why

Historically, web automation relied on reading a website's background code (DOM manipulation) or reverse-engineering hidden APIs. In the current digital landscape, trillion-dollar AI systems employed by social networks instantly detect and ban headless browsers and scraper scripts. 

It does not read code; it reads pixels. It does not send HTTP requests to internal endpoints; it physically commands the operating system's hardware mouse and keyboard drivers. 
1. **The Eyes:** Vertex AI (Gemini) takes a screenshot of the virtual monitor and returns the exact [X, Y] pixel coordinates of the interface elements we want to interact with.
2. **The Hands:** We utilize low-level C-bindings (`RobotGo`) to emulate human kinematics—moving the mouse along randomized Bezier curves and typing with millisecond, human-like cadences.

## Architecture: The S66 7-Layer OSI Model

To help conceptualize this invisible "headless" terminal bot, we map the VLA architecture to a 7-layer OSI structure:

| Layer | Name | Component | Function in S66 Reachout Bot |
|---|---|---|---|
| **7** | **Application (Personality)** | `SKILL.md` | The highest layer. Dictates the bot's "soul", persona, tone, and specific Atlanta System Films outreach instructions. |
| **6** | **Presentation (Parsing)** | `pkg/optics/vision.go` | Translates the unstructured Vertex JSON response into precise `[Y, X]` float arrays. |
| **5** | **Session (FSM Loop)** | `cmd/s66-bot/main.go` | The Infinite State Machine that manages the active DM session (Wait -> Screenshot -> Reason -> Click -> Type). |
| **4** | **Transport (API / Token)** | `gcloud ADC` | The secure, encrypted tunnel that transports images and text to Google Cloud via your Application Default Credentials. |
| **3** | **Network (Vertex Cloud)** | `genai.Client` | Gemini 1.5 Pro multimodal processing. This is the "Brain" that does the heavy lifting of understanding the screen. |
| **2** | **Data Link (Memory)** | `pkg/pipeline` | Firestore / SQLite tracking. Ensures we don't message the same user twice. |
| **1** | **Physical (Actuation)** | `pkg/hardware/motor.go` | Raw OS-level C-bindings. Uses `robotgo` to capture physical VRAM bytes and actuate literal hardware mouse/keyboard events. |

Because Layer 1 completely bypasses the browser's DOM (HTML/CSS), Instagram cannot block it. It is physically simulating a human being.

## Visual Debugging
Because the bot runs in the background without a UI dashboard, we built **Observability** into Layer 1. 
Every time the bot takes a screenshot, it saves it to the root folder as `debug_vision.jpg`. You can click this file to literally "see" what the bot is seeing at any given moment.

To the platform's security algorithms, system-dm-bot is indistinguishable from a System Films A&R representative sitting at a computer in Atlanta, reading a screen, and physically clicking a mouse.

---

## 🌐 Universal Laptop Setup (God Mode)

Because this bot physically controls the operating system, it has a strict setup process. If you are a teammate pulling this repo to a new laptop, you must follow these steps perfectly.

### Step 1: The "C Thing" (TDM-GCC)
Because the bot uses `RobotGo` to emulate hardware inputs, it requires C-bindings. Standard Golang cannot compile this natively on Windows.
1. Download and install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/).
2. You **must** add the TDM-GCC `bin` folder to your Windows System Environment `PATH` (e.g., `C:\TDM-GCC-64\bin`).

### Step 2: Google Cloud CLI & Billing
The bot sends images to Vertex AI (Gemini 2.5 Flash). You need a Google Cloud account with an active billing profile.
1. Install the [Google Cloud CLI](https://cloud.google.com/sdk/docs/install).
2. Open a terminal and log in:
   ```bash
   gcloud auth application-default login
   ```
3. **CRITICAL:** You must tell Google which project is paying for the API calls. Run this command to set the Quota Project:
   ```bash
   # Use matrix-esa-production for testing, or system-dm-bot for production
   gcloud auth application-default set-quota-project matrix-esa-production
   ```

### Step 3: Project Configuration Switch
In `cmd/test-vision/main.go` and `cmd/s66-bot/main.go`, ensure the `NewVisionClient` string matches your target Google Cloud Project ID. 
- *Testing:* `"matrix-esa-production"`
- *Production:* `"system-dm-bot"`

### Step 4: Compiling & Testing
Always test the "Eyes" and "Hands" independently before running the main bot loop.

**Test the Hands (Physical Actuation):**
```bash
$env:CGO_ENABLED="1"
go build -buildmode=exe -ldflags="-s -w" -o test_hands.exe ./cmd/test-hardware/main.go
.\test_hands.exe
```
*(Warning: The moment you press enter, take your hand off the mouse).*

**Test the Eyes (Vertex AI Vision):**
```bash
$env:CGO_ENABLED="1"
go build -buildmode=exe -ldflags="-s -w" -o test_vision.exe ./cmd/test-vision/main.go
.\test_vision.exe
```
*(Check the root folder for `debug_vision.jpg` to see what the bot saw).*

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
