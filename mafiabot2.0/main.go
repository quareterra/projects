package main

import (
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Room struct {
	id int64
}

type Player struct {
	id   int64
	room int64
}

var (
	rooms   []Room
	players []Player
)

type State uint

const (
	IdleState State = iota
	EnteringRoomIdState
	InRoomState
)

func getMessageID() {
}

func main() {
	/* setup bot */
	_, err := tgbotapi.NewBotAPI(os.Getenv("7401254673:AAGR-g_Ur41t9d1DgjRzK7uvxhmr7CSCBVs"))
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		playerId := update.FromChat().ID
		player := getPlayerById
	}
}
