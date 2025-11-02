package ports

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lucasm0ta/change-detector/core"
)

//
type TelegramBot struct {
	//
	TelegramAccessToken string

	//
	watcherService *core.WatcherService
}

func buildWatchRequest(url string, chatId string) (*core.WatchRequest, error) {
	channel := core.NewChannelInfo("telegram", chatId)
	watchequest, err := core.NewWatchRequest(channel, url)
	if err != nil {
		return nil, err
	}
	return watchequest, nil
}

func NewTelegramBot(telegramAccessToken string, watcherService *core.WatcherService) *TelegramBot {
	telegramBot := new(TelegramBot)
	telegramBot.TelegramAccessToken = telegramAccessToken
	telegramBot.watcherService = watcherService
	return telegramBot
}

func (telegram *TelegramBot) Start() {
	bot, err := tgbotapi.NewBotAPI(telegram.TelegramAccessToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Set up the callback for page changes
	telegram.watcherService.OnChangeCallback = func(reponse *core.WatchResponse) {
		chatID, _ := strconv.ParseInt(reponse.ChannelID, 10, 64)
		msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("🔔 Change detected!\n%s\n\n\n%s", reponse.URL.String(), reponse.Diff))
		if _, err := bot.Send(msg); err != nil {
			log.Printf("Error sending notification: %v", err)
		}
	}

	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "watch":
				url := update.Message.CommandArguments()
				msg.Text = telegram.Watch(url, fmt.Sprint(update.Message.Chat.ID))
			case "list":
				var result strings.Builder
				for i, watched := range telegram.watcherService.GetWatched() {
					result.WriteString(fmt.Sprintf("%d. %s\n", i+1, watched.Request.URL.String()))
				}
				msg.Text = result.String()
			case "remove":
				msg.Text = "removed"
			case "help":
				msg.Text = "can't help U, sorry :/"
			default:
				msg.Text = "Sorry, your command "
			}
		} else {
			msg.Text = `
			Sorry, I Can only understand the following commands:
			/watch
			/list
			/remove
			`
		}
		bot.Send(msg)
	}
}

func (telegram *TelegramBot) Watch(url string, chatID string) string {
	if url == "" {
		return "pass a URL please"
	}
	watchRequest, err := buildWatchRequest(url, chatID)
	if err != nil {
		log.Printf("Error building watch request: %v", err)
		return "oooops could not watch this"
	}
	err = telegram.watcherService.Register(watchRequest)
	if err == nil {
		return "now watching"
	} else {
		log.Printf("Error watching request: %v", err)
		return "oooops could not watch this"
	}
}
