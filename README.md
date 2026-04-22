# System DM Bot (VLA Sovereign Agent)

![Architecture](debug_vision.jpg)

The **System DM Bot** is a Vision-Language-Action (VLA) Sovereign Agent framework designed for universal, automated outreach. It autonomously navigates desktop web UIs, visually hunts for inbound leads using Computer Vision, evaluates conversation context using Cognitive AI, and dynamically generates personalized outreach pitches.

Unlike traditional headless scrapers, DOM-dependent bots, or API-based automation (which are heavily rate-limited and easily detected), this architecture relies entirely on physical OS-level kinematic inputs. It physically moves the mouse and types on the keyboard, making it highly resilient to standard bot-detection systems.

## System Architecture

The bot operates on a robust Finite State Machine (FSM) loop powered by three core layers:

1. **Optics Layer (`pkg/optics`)**: Uses Google Vertex AI (Gemini 2.5 Flash) to analyze spatial UI coordinates from raw screen captures and perform deep cognitive evaluations of chat history.
2. **Motor Layer (`pkg/hardware`)**: Uses `robotgo` to hijack physical OS-level mouse and keyboard drivers. It simulates human-like kinematic movements (e.g., bezier curve mouse paths, randomized typing delays) to avoid detection heuristics.
3. **Cognitive Engine (`SKILL.md`)**: A flexible, markdown-based persona file injected into the prompt pipeline. This acts as the "brain" of the agent, dictating strict business rules, tone, and the exact constraints for generating personalized responses.

---

## Universal Laptop Setup (Operator Deployment)

The System DM Bot is designed for scalable, secure deployment across organizational teams using centralized **Google Cloud Platform (GCP)** architecture.

> **[IMPORTANT] Security Standard**
> **DO NOT** distribute raw API keys or Service Account JSON files to team members or end-users. 
> This framework is built on GCP Identity and Access Management (IAM). Billing and infrastructure are centrally controlled by the Organization Administrator. Operators simply log in via their corporate Google accounts. If an operator leaves, the Admin revokes their IAM role, immediately terminating their bot's cognitive engine.

If you are a team operator installing this on your local laptop, follow these exact steps:

### 1. Install Prerequisites
- **[Go (Golang)](https://go.dev/dl/)**: Required to compile the engine.
- **[TDM-GCC](https://jmeubank.github.io/tdm-gcc/)**: A C/C++ compiler required for the CGO hardware drivers (`robotgo`). Add it to your Windows `PATH`.
- **[Google Cloud CLI](https://cloud.google.com/sdk/docs/install)**: Required for secure enterprise authentication.

### 2. Authenticate with Google Cloud (IAM)
Open your terminal and authenticate using your designated corporate Google account. This securely bridges your local hardware to the organization's centralized billing cloud:

```bash
gcloud auth application-default login --quota-project=YOUR_GCP_PROJECT_ID
```
*(Note: Ensure your Organization Admin has granted your email the `Vertex AI User` IAM role before proceeding, otherwise the cognitive engine will fail).*

### 3. Compile the Executable
Clone the repository and build the binary on your local machine.

```powershell
git clone <REPOSITORY_URL>
cd system-dm-bot
go build -ldflags="-s -w" -o system-dm-bot.exe ./cmd/s66-bot/main.go
```
*(Note: The build path may vary based on your organization's specific white-label entry point, e.g., `./cmd/bot/main.go`).*

---

## Execution Manual

Before running the agent, you must manually align your physical desktop environment. The bot does not interact with hidden windows—it must physically "see" the screen.

1. **Hardware Setup**: Move your target web application (e.g., Instagram Web, LinkedIn) to your **Primary Monitor**. Ensure the window is fully maximized.
2. **UI Setup**: Navigate to the required application state (e.g., the primary DM Inbox page).
3. **Launch the Engine**: Open your terminal window. Keep the terminal visible (ideally docked to the bottom or side of the screen) and run:
   ```powershell
   .\system-dm-bot.exe
   ```
4. **Hands Off**: The moment you press Enter, take your hands off the physical mouse and keyboard. The Sovereign Agent is now driving your machine.

### The Kill Switch (Emergency Stop)
Because the bot hijacks your physical OS drivers and runs on an infinite state loop, it will not stop until you intervene. 
If the bot misbehaves or you need to shut it down:
**Drag your mouse over to the terminal window and press `CTRL + C` on your physical keyboard.** This will instantly terminate the CGO hardware process.
