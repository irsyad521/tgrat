# tgrat — Telegram C2 Agent

## Disclosure

> **⚠️ FOR RESEARCH & EDUCATIONAL PURPOSES ONLY**
>
> This project was created strictly for:
>
> * **Detection Engineering** — building and validating behavioral C2 detection
> * **Security Research** — analyzing Living Off Trusted Services (LOTS) traffic patterns
> * **Proof of Concept** — simulating beaconing behavior in isolated lab environments
>
> This project is intended solely for use in environments that you own or have explicit authorization to test.
>
> Any unauthorized use against third-party systems may be illegal and is strictly discouraged.
>
> The author assumes **no responsibility** for any misuse of this project. Deploy and use it only for legitimate security research, defensive testing, and educational purposes.

---

## Overview

`tgrat` is a Telegram-based C2 agent simulation written in Go, designed as a **Proof of Concept** for detection engineering research against **LOTS (Living Off Trusted Services)** attacks.

The agent uses Telegram's Bot API long polling as its command channel — a technique observed in real-world threat actor operations where legitimate trusted services are abused as C2 infrastructure to evade reputation-based detection.

Beacon behavior is configured using **Beacon Profiles**, allowing different polling characteristics without modifying the application logic.

Available profiles include:

* **Baseline** — fixed interval, high beacon consistency (baseline detection scenario)
* **Small Jitter** — randomized interval with low variance
* **Medium Jitter** — randomized interval with moderate variance
* **Large Jitter** — randomized interval with high variance

---

## Project Structure

```text
tgrat/
├── .env                # Telegram credentials (local, not versioned)
├── .env.example        # Example configuration template
├── .gitignore          # Git ignore rules
├── build.sh            # Build script for standalone binary
├── commands.go         # Command dispatcher & shell execution
├── config.go           # Configuration & beacon profiles
├── telegram.go         # Telegram Bot API methods
├── main.go             # Main polling loop
├── go.mod              # Go module definition
├── go.sum              # Dependency checksums
└── README.md           # This file
```

---

## Prerequisites

* Go 1.21+
* Telegram Bot Token (from [@BotFather](https://t.me/BotFather))
* Target Chat ID
* Isolated lab environment (VM or container)

---

## Setup

### 1. Create a Telegram Bot

```text
1. Open Telegram → search @BotFather
2. Send /newbot → follow the instructions → copy the TOKEN
3. Send any message to your bot, then open:
   https://api.telegram.org/bot<TOKEN>/getUpdates
4. Copy the chat.id value from the JSON response
```

### 2. Configure Credentials

Create `.env` file with your Telegram credentials:

```bash
cp .env.example .env
```

Edit `.env` with your credentials:

```dotenv
TELEGRAM_BOT_TOKEN=your_bot_token_here
TELEGRAM_CHAT_ID=your_chat_id_here
BEACON_PROFILE=baseline
```

> **Note:** `.env` is used only for building the binary. Once built, the binary is completely standalone and does not require the `.env` file.

### 3. Beacon Profiles

Available beacon profiles that can be embedded at build time:

| Profile | Description | Example Poll Interval* |
| --- | --- | --- |
| `baseline` | Fixed polling interval with no jitter. Produces consistent beacon traffic for baseline detection validation. | `61s → 61s → 61s → 61s` |
| `small` | Small randomized delay while maintaining recognizable pattern. Good for basic jitter tolerance testing. | `59s → 64s → 61s → 66s` |
| `medium` | Moderate random delay simulating real-world beaconing behavior. | `48s → 73s → 58s → 82s` |
| `large` | Maximum randomized delay producing highly irregular intervals. Useful for robust detection evaluation. | `35s → 97s → 54s → 76s` |

*Example intervals are illustrative only. Actual intervals depend on the beacon profile configuration.*

### 4. Build the Agent

The agent is built as a **standalone binary** with all configuration (credentials, beacon profile) embedded at build time. Once built, the binary requires no external files and can be deployed anywhere.

#### Quick Start (Recommended)

```bash
# Build with .env credentials and default profile (baseline)
./build.sh

# Build with .env credentials and custom profile
BEACON_PROFILE=medium ./build.sh

# Build with all custom values
TELEGRAM_BOT_TOKEN=xxx TELEGRAM_CHAT_ID=yyy BEACON_PROFILE=large ./build.sh
```

#### Manual Build

Build for different platforms, embedding all configuration:

**Linux:**
```bash
go build -ldflags "-X main.TOKEN=<TOKEN> -X main.CHAT_ID=<CHAT_ID> -X main.BEACON_PROFILE=small" -o agent
```

**Windows:**
```bash
GOOS=windows GOARCH=amd64 go build -ldflags "-X main.TOKEN=<TOKEN> -X main.CHAT_ID=<CHAT_ID> -X main.BEACON_PROFILE=medium" -o agent.exe
```

**macOS:**
```bash
GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.TOKEN=<TOKEN> -X main.CHAT_ID=<CHAT_ID> -X main.BEACON_PROFILE=large" -o agent_macos
```

#### Development/Testing

For testing without building:

```bash
go mod tidy
go run .
```

This requires `.env` file in the working directory.

#### Deployment

Once built, the binary is fully portable and standalone:

```bash
# Copy to any system and run directly
cp agent /tmp/
cd /tmp
./agent  # Runs with embedded configuration
```

---

## Commands

Send commands via the Telegram bot chat:

| Command | Description |
| ----------------- | --------------------------------------------------------------------------------------------- |
| `help` | Show help menu |
| `info` | Display system information (OS, architecture, hostname, CPU, current user, working directory) |
| `location` | Retrieve the public IP address and geolocation |
| `jobs` | List all running background shell jobs |
| `kill <id>` | Terminate a background shell job by its ID |
| `cd <dir>` | Change the current working directory |
| `cd ..` | Move to the parent directory |
| `download <file>` | Upload a local file from the agent to Telegram |
| `get <url>` | Download a file from a URL to the agent |
| `shell <cmd>` | Execute a shell command |
| `shell <cmd> &` | Execute a shell command in the background |

---

## Lab Setup

```text
┌─────────────────────────────────────────┐
│           Isolated Network              │
│                                         │
│  ┌──────────┐         ┌──────────────┐  │
│  │ Victim VM│         │ Detection VM │  │
│  │  agent   │         │  (your NDR)  │  │
│  └────┬─────┘         └──────────────┘  │
│       │          mirrored/tap traffic   │
└───────┼─────────────────────────────────┘
        │ outbound HTTPS
        ▼
   api.telegram.org:443
```

Run the agent on the Victim VM, capture traffic on the Detection VM using your preferred NDR stack, and observe the behavioral patterns produced by the selected beacon profile.

---

## References

* [LOTS Project — Living Off Trusted Sites](https://lots-project.com)
* [RITA — Real Intelligence Threat Analytics](https://github.com/activecm/rita)
* [Zeek Network Security Monitor](https://zeek.org)
* [Telegram Bot API](https://core.telegram.org/bots/api)