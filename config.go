package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Telegram Credentials
// ─────────────────────────────────────────────────────────────────────────────

var (
	TOKEN   = ""
	CHAT_ID = ""
)

// ─────────────────────────────────────────────────────────────────────────────
// Beacon Profiles
// ─────────────────────────────────────────────────────────────────────────────

type BeaconProfile struct {
	Name string

	PollTimeout time.Duration

	BaseSleep time.Duration

	JitterEnabled bool
	JitterMin     time.Duration
	JitterMax     time.Duration

	BackoffEnabled bool
	BackoffBase    time.Duration
	BackoffMax     time.Duration
}

var Profiles = map[string]BeaconProfile{

	"baseline": {
		Name:           "Baseline",
		PollTimeout:    60 * time.Second,
		BaseSleep:      1 * time.Second,
		JitterEnabled:  false,
		JitterMin:      0,
		JitterMax:      0,
		BackoffEnabled: true,
		BackoffBase:    5 * time.Second,
		BackoffMax:     5 * time.Minute,
	},

	"small": {
		Name:           "Small Jitter",
		PollTimeout:    60 * time.Second,
		BaseSleep:      1 * time.Second,
		JitterEnabled:  true,
		JitterMin:      1 * time.Second,
		JitterMax:      10 * time.Second,
		BackoffEnabled: true,
		BackoffBase:    5 * time.Second,
		BackoffMax:     5 * time.Minute,
	},

	"medium": {
		Name:           "Medium Jitter",
		PollTimeout:    60 * time.Second,
		BaseSleep:      1 * time.Second,
		JitterEnabled:  true,
		JitterMin:      5 * time.Second,
		JitterMax:      30 * time.Second,
		BackoffEnabled: true,
		BackoffBase:    5 * time.Second,
		BackoffMax:     5 * time.Minute,
	},

	"large": {
		Name:           "Large Jitter",
		PollTimeout:    60 * time.Second,
		BaseSleep:      1 * time.Second,
		JitterEnabled:  true,
		JitterMin:      15 * time.Second,
		JitterMax:      60 * time.Second,
		BackoffEnabled: true,
		BackoffBase:    5 * time.Second,
		BackoffMax:     5 * time.Minute,
	},
}

var Profile BeaconProfile

func LoadProfile() error {

	name := strings.ToLower(os.Getenv("BEACON_PROFILE"))

	if name == "" {
		name = "baseline"
	}

	p, ok := Profiles[name]

	if !ok {
		return fmt.Errorf(
			"invalid beacon profile '%s' (available: baseline, small, medium, large)",
			name,
		)
	}

	Profile = p

	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Telegram
// ─────────────────────────────────────────────────────────────────────────────

const (
	BaseURLPrefix  = "https://api.telegram.org/bot"
	MaxMessageSize = 4000
)

var HTTPClient = &http.Client{
	Timeout: 90 * time.Second,
}

func BaseURL() string {
	return BaseURLPrefix + TOKEN
}
