#!/bin/bash

# Save CLI arguments (they take precedence over .env)
CLI_TOKEN="$TELEGRAM_BOT_TOKEN"
CLI_CHAT_ID="$TELEGRAM_CHAT_ID"
CLI_PROFILE="$BEACON_PROFILE"

# Load from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# CLI args override .env
[ -n "$CLI_TOKEN" ] && TELEGRAM_BOT_TOKEN="$CLI_TOKEN"
[ -n "$CLI_CHAT_ID" ] && TELEGRAM_CHAT_ID="$CLI_CHAT_ID"
[ -n "$CLI_PROFILE" ] && BEACON_PROFILE="$CLI_PROFILE"

# Set defaults if not defined
: ${TELEGRAM_BOT_TOKEN:=your_bot_token}
: ${TELEGRAM_CHAT_ID:=your_chat_id}
: ${BEACON_PROFILE:=baseline}

# Build with embedded configuration
go build -ldflags "-X main.TOKEN=$TELEGRAM_BOT_TOKEN -X main.CHAT_ID=$TELEGRAM_CHAT_ID -X main.BEACON_PROFILE=$BEACON_PROFILE" -o agent

echo "✓ Standalone binary built successfully: ./agent"
echo "  Profile: $BEACON_PROFILE"
