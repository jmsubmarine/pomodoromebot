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
		send(bot, chatID, `Привет! Я помогу тебе сосредоточиться 🍅

Напиши /runpomodoro, чтобы запустить серию таймеров по технике Помидоро. После каждого 25-минутного раунда будет 5 минут перерыва.

Если тебе нужна помощь — напиши /help.`)

	case "help":
		send(bot, chatID, `Что я умею:

/runpomodoro — начать серию таймеров  
/stop — остановить текущую сессию  
/status — узнать статус текущей сессии  
/music — получить музыку для фона`)

	case "runpomodoro":
		if session, ok := Store.Get(chatID); ok && session.Timer != nil {
			send(bot, chatID, "У тебя уже идёт сессия. Напиши /stop, если хочешь начать заново.")
			return
		}

		Store.Set(chatID, &PomodoroSession{AwaitingInput: true})
		send(bot, chatID, "Сколько раундов ты хочешь сделать? Введи число от 1 до 10:")

	case "stop":
		stopSession(bot, chatID)

	case "status":
		handleStatus(bot, chatID)

	case "music":
		sendHTML(bot, chatID,
			`Вот список рекомендаций с музыкой для фона:

<a href="https://youtu.be/t3LCXpKI9K0?si=27yfd61dEv82lgVj">Redwood Resonance</a>  
<a href="https://youtu.be/wIBnaNuhuCQ?si=dTYX0vD-3ZLqUi7e">ASMR New York Library</a>  
<a href="https://youtu.be/ecechHEtkYU?si=uzsf6K7IV7WKvtVl">quiet mornings, slowly waking up to the smell of fresh coffee</a>  
<a href="https://youtu.be/tFAjJsqdO_A?si=cpy3BS__3J9-6Fjz">Harry Potter Chill Music ~ Hogwarts Library</a>`)
	default:
		send(bot, chatID, "Неизвестная команда.")
	}
}

func handleUserInput(bot *tgbotapi.BotAPI, chatID int64, text string) {
	session, ok := Store.Get(chatID)
	if !ok || !session.AwaitingInput {
		return
	}

	n, err := strconv.Atoi(text)
	if err != nil || n < MinRounds || n > MaxRounds {
		send(bot, chatID, fmt.Sprintf("Пожалуйста, введи число от %d до %d.", MinRounds, MaxRounds))
		return
	}

	session.AwaitingInput = false
	session.TotalRounds = n
	session.CurrentRound = 1
	session.Phase = PhaseWork
	session.StartTime = time.Now()

	Store.Set(chatID, session) // Обновляем сессию после изменения

	startWorkRound(bot, chatID)
}

func startWorkRound(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		return
	}

	session.Phase = PhaseWork
	session.StartTime = time.Now()

	send(bot, chatID, fmt.Sprintf("Раунд %d: начинаем работу на 25 минут!", session.CurrentRound))

	session.Timer = time.AfterFunc(DefaultWork, func() {
		send(bot, chatID, "⏰ Время перерыва! 5 минут отдыха.")
		startBreak(bot, chatID)
	})

	Store.Set(chatID, session) // Обновляем сессию после изменения
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
			send(bot, chatID, "🎉 Все раунды завершены! Отличная работа!")
			Store.Delete(chatID)
			return
		}

		send(bot, chatID, fmt.Sprintf("🔔 Перерыв окончен. Начинаем раунд %d!", session.CurrentRound))
		startWorkRound(bot, chatID)
	})

	Store.Set(chatID, session) // Обновляем сессию после изменения
}

func handleStatus(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		send(bot, chatID, "Нет активной сессии.")
		return
	}

	var duration time.Duration
	switch session.Phase {
	case PhaseWork:
		duration = DefaultWork
	case PhaseBreak:
		duration = DefaultBreak
	default:
		send(bot, chatID, "Неизвестная фаза.")
		return
	}

	remaining := duration - time.Since(session.StartTime)
	if remaining < 0 {
		remaining = 0
	}

	msg := fmt.Sprintf(
		"📊 Раунд %d из %d\nФаза: %s\nОставшееся время: %v",
		session.CurrentRound, session.TotalRounds, session.Phase, remaining.Round(time.Second),
	)
	send(bot, chatID, msg)
}

func stopSession(bot *tgbotapi.BotAPI, chatID int64) {
	session, ok := Store.Get(chatID)
	if !ok {
		send(bot, chatID, "Нет активной сессии.")
		return
	}

	if session.Timer != nil {
		session.Timer.Stop()
	}
	Store.Delete(chatID)
	send(bot, chatID, "⏹ Сессия остановлена.")
}

func send(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if _, err := bot.Send(msg); err != nil {
		log.Printf("ошибка отправки: %v", err)
	}
}

func sendHTML(bot *tgbotapi.BotAPI, chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "HTML"
	if _, err := bot.Send(msg); err != nil {
		log.Printf("ошибка отправки (HTML): %v", err)
	}
}
