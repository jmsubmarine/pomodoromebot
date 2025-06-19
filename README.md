# Pomodoromebot 🍅

Lightweight and user-friendly Pomodoro timer for Telegram developed in Go.  
Helps you manage your time and tasks effectively right inside the messenger.

---

## Features

- Start Pomodoro sessions with customizable number of rounds  
- Work intervals of 25 minutes and 5-minute breaks by default  
- Commands for starting, stopping, checking status, and music recommendations  
- Session management with Telegram chat context  
- Simple and clean user interaction  

---

## Commands

- `/runpomodoro` — Start a Pomodoro session  
- `/stop` — Stop the current session  
- `/status` — Check the status of the current session  
- `/music` — Get recommended background music links  
- `/help` — Show help message  

---

## Installation

1. Clone the repository  
   ```bash
   git clone https://github.com/jmsubmarine/pomodoromebot.git
   cd pomodoromebot

2. Set up environment variable with your Telegram bot token

export BOT_TOKEN="your-telegram-bot-token"

3. Build and run the bot

go build -o pomodoromebot
./pomodoromebot