package bot

import (
	"fmt"
	"log"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	MaxRounds    = 10
	MinRounds    = 1
	DefaultWork  = 25 * time.Minute
	DefaultBreak = 5 * time.Minute
)

func HandleUpdate(bot *tgbotapi.BotAPI, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID

	if update.Message.IsCommand() {
		handleCommand(bot, chatID, update.Message.Command())
		return
	}

	handleUserInput(bot, chatID, update.Message.Text)
}

func handleCommand(bot *tgbotapi.BotAPI, chatID int64, command string) {
	switch command {
	case "start":
		send(bot, chatID, `Hi! I'm here to help you focus 🍅

Send /runpomodoro to start a series of Pomodoro timers. After each 25-minute work round, you'll get a 5-minute break.

If you need help, type /help.`)

	case "help":
		send(bot, chatID, `What I can do:

/runpomodoro — start a Pomodoro timer session  
/stop — stop the current session  
/status — check the status of the current session  
/music — get background music suggestions`)

	case "runpomodoro":
		if session, ok := Store.Get(chatID); ok && session.Timer != nil {
			send(bot, chatID, "You already have an active session. Send /stop if you want to start over.")
			return
		}

		Store.Set(chatID, &PomodoroSession{AwaitingInput: true})
		send(bot, chatID, "How many rounds would you like to do? Enter a number from 1 to 10:")

	case "stop":
		stopSession(bot, chatID)

	case "status":
		handleStatus(bot, chatID)

	case "music":
		sendHTML(bot, chatID,
			`Here’s a list of background music recommendations:

<a href="https://youtu.be/t3LCXpKI9K0?si=27yfd61dEv82lgVj">Redwood Resonance</a>  
<a href="https://youtu.be/wIBnaNuhuCQ?si=dTYX0vD-3ZLqUi7e">ASMR New York Library</a>  
<a href="https://youtu.be/ecechHEtkYU?si=uzsf6K7IV7WKvtVl">quiet mornings, slowly waking up to the smell of fresh coffee</a>  
<a href="https://youtu.be/tFAjJsqdO_A?si=cpy3BS__3J9-6Fjz">Harry Potter Chill Music ~ Hogwarts Library</a>`)
	default:
		send(bot, chatID, "Unknown command.")
	}
}

func handleUserInput(bot *tgbotapi.BotAPI, chatID int64, text string) {
	session, ok := Store.Get(chatID)
	if !ok || !session.AwaitingInput {
		return
	}

	n, err := strconv.Atoi(text)
	if err != nil || n < MinRounds || n > MaxRounds {
		send(bot, chatID, fmt.Sprintf("Please enter a number between %d and %d.", MinRounds, MaxRounds))
		return
	}

	session.AwaitingInput = false
	session.TotalRounds = n
	session.CurrentRound = 1
	session.Phase = PhaseWork
	session.StartTime = time.Now()

	Store.Set(chatID, session)

	startWorkRound(bot, chatID)
}

func startWorkRound(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		return
	}

	session.Phase = PhaseWork
	session.StartTime = time.Now()

	send(bot, chatID, fmt.Sprintf("Round %d: time to focus for 25 minutes!", session.CurrentRound))

	session.Timer = time.AfterFunc(DefaultWork, func() {
		send(bot, chatID, "⏰ Break time! Take 5 minutes to rest.")
		startBreak(bot, chatID)
	})

	Store.Set(chatID, session)
}

func startBreak(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		return
	}

	session.Phase = PhaseBreak
	session.StartTime = time.Now()

	session.Timer = time.AfterFunc(DefaultBreak, func() {
		session.CurrentRound++

		if session.CurrentRound > session.TotalRounds {
			send(bot, chatID, "🎉 All rounds are complete! Great job!")
			Store.Delete(chatID)
			return
		}

		send(bot, chatID, fmt.Sprintf("🔔 Break is over. Starting round %d!", session.CurrentRound))
		startWorkRound(bot, chatID)
	})

	Store.Set(chatID, session)
}

func handleStatus(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		send(bot, chatID, "No active session.")
		return
	}

	var duration time.Duration
	switch session.Phase {
	case PhaseWork:
		duration = DefaultWork
	case PhaseBreak:
		duration = DefaultBreak
	default:
		send(bot, chatID, "Unknown phase.")
		return
	}

	remaining := duration - time.Since(session.StartTime)
	if remaining < 0 {
		remaining = 0
	}

	msg := fmt.Sprintf(
		"📊 Round %d of %d\nPhase: %s\nTime remaining: %v",
		session.CurrentRound, session.TotalRounds, session.Phase, remaining.Round(time.Second),
	)
	send(bot, chatID, msg)
}

func stopSession(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		send(bot, chatID, "No active session.")
		return
	}

	if session.Timer != nil {
		session.Timer.Stop()
	}
	Store.Delete(chatID)
	send(bot, chatID, "⏹ Session stopped.")
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send error: %v", err)
	}
}

func sendHTML(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("send error (HTML): %v", err)
	}
}
