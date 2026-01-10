package main

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Room struct {
	id int64
}

type Player struct {
	id    int64
	name  string
	room  int64
	state State
	role  Role
}

var (
	rooms   []Room
	players []Player
)

type (
	State uint
	Role  uint
)

const (
	IdleState State = iota
	JoiningRoomState
	InRoomIdleState
	InRoomReadyState
	InGameState
)

const (
	CivilianRole Role = iota
	MafiaRole
)

const (
	BtnCreate = "create_room"
	BtnJoin   = "join_room"
	BtnReady  = "ready"
	BtnLeave  = "leave"
)

func handleCallback(bot *tgbotapi.BotAPI, callback *tgbotapi.CallbackQuery) {
	bot.Request(tgbotapi.NewCallback(callback.ID, ""))
	response := tgbotapi.NewMessage(callback.Message.Chat.ID, "")
	playerptr := getOrCreatePlayerById(callback.Message.Chat.ID, callback.From.UserName)

	switch callback.Data {
	case BtnCreate:
		if playerptr.state == IdleState {
			room := createRoom(playerptr)
			response.Text = fmt.Sprintf("room ID: %d", room.id)
			response.ReplyMarkup = getKeyboard(playerptr.state)
		}

	case BtnJoin:
		if playerptr.state == IdleState {
			playerptr.state = JoiningRoomState
			response.Text = "enter room id"
			response.ReplyMarkup = getKeyboard(playerptr.state)
		}

	case BtnLeave:
		if playerptr.state == InRoomIdleState || playerptr.state == InRoomReadyState {
			sendMessageToOwner(bot, getOwnerByPLayer(playerptr), "leave", playerptr)
			leaveRoom(playerptr)
			response.Text = "you left the room"
			response.ReplyMarkup = getKeyboard(playerptr.state)
		} else if playerptr.state == JoiningRoomState {
			playerptr.state = IdleState
			response.Text = "canceled"
			response.ReplyMarkup = getKeyboard(playerptr.state)
		}

	case BtnReady:
		if playerptr.state == InRoomIdleState {
			playerptr.state = InRoomReadyState
			ownerptr := getOwnerByPLayer(playerptr)
			sendMessageToOwner(bot, ownerptr, "ready", playerptr)
			if canGameStart(ownerptr) {
				startGameByOwner(bot, ownerptr)
			}
		}
	}

	if response.Text != "" {
		_, err := bot.Send(response)
		if err != nil {
			log.Print(err)
		}
	}
}

func handleMessage(bot *tgbotapi.BotAPI, message *tgbotapi.Message) {
	response := tgbotapi.NewMessage(message.Chat.ID, "")
	playerptr := getOrCreatePlayerById(message.Chat.ID, message.From.UserName)

	switch playerptr.state {

	case IdleState:
		response.Text = "create or join room"
		response.ReplyMarkup = getKeyboard(playerptr.state)

	case JoiningRoomState:
		if addPlayerToRoom(playerptr, message.Text) {
			sendMessageToOwner(bot, getOwnerByPLayer(playerptr), "join", playerptr)
			response.ReplyMarkup = getKeyboard(playerptr.state)
			response.Text = fmt.Sprintf("you joined room %s", message.Text)
		} else {
			response.Text = "incorrect try again"
			response.ReplyMarkup = getKeyboard(playerptr.state)
		}

	case InRoomIdleState:
		response.Text = "ready to play?"
		response.ReplyMarkup = getKeyboard(playerptr.state)

	case InRoomReadyState:
		response.Text = "wait for game to start"
		response.ReplyMarkup = getKeyboard(playerptr.state)
	}

	if response.Text != "" {
		_, err := bot.Send(response)
		if err != nil {
			log.Print(err)
		}
	}

}

func main() {
	bot, err := tgbotapi.NewBotAPI("7401254673:AAGR-g_Ur41t9d1DgjRzK7uvxhmr7CSCBVs")
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.CallbackQuery != nil {
			handleCallback(bot, update.CallbackQuery)
		} else if update.Message != nil {
			handleMessage(bot, update.Message)
		}

		fmt.Println("rooms:", rooms)
		fmt.Println("players:", players)
	}
}
