package main

import (
	"log"
	"os"

	"github.com/lucasm0ta/change-detector/core"
	"github.com/lucasm0ta/change-detector/ports"
)

func main() {
	watcherService := core.NewWatcherService()
	go watcherService.Start()

	telegram_access_token := os.Getenv("TELEGRAM_ACCESS_TOKEN")
	if len(telegram_access_token) == 0 {
		log.Panic("Missing evironment variable TELEGRAM_ACCESS_TOKEN")
	}
	telegram := ports.NewTelegramBot(telegram_access_token, watcherService)
	telegram.Start()
}
