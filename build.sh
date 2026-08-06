#!/bin/bash

# Load from .env if it exists
if [ -f .env ]; then
    export $(grep -v '^#' .env | xargs)
fi

# Build with embedded credentials
go build -ldflags "-X main.TOKEN=$TELEGRAM_BOT_TOKEN -X main.CHAT_ID=$TELEGRAM_CHAT_ID" -o agent

echo "✓ Standalone binary built successfully: ./agent"
