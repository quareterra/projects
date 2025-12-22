package main

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Room struct {
	id int64
}

type State uint

const (
	IdleState State = iota
	EnteringRoomIdState
	InRoomState
)

type ButtonData string

const (
	CreateRoomButtonData ButtonData = "create_room"
	JoinRoomButtonData   ButtonData = "join_room"
)

type User struct {
	id     int64
	roomId int64
	state  State
}

var users []User
var rooms []Room

func createRoom(userptr *User) (txt string) {
	if (*userptr).state != IdleState {
		txt = "not idle state"
		return
	}

	rooms = append(rooms, Room{(*userptr).id})
	(*userptr).state = InRoomState
	(*userptr).roomId = (*userptr).id
	txt = fmt.Sprintf("room ID = %d", (*userptr).id)
	return
}

func deleteRoom(userptr *User) (txt string) {
	if (*userptr).state != InRoomState || !isUserOwner(userptr) {
		txt = "not room owner"
		return
	}

	for i := range users {
		if users[i].roomId == (*userptr).id {
			users[i].roomId = 0
			users[i].state = IdleState
			txt = "room deleted"
		}
	}
	return
}

func joinRoomButtonPressedAction(userptr *User) (txt string) {
	if (*userptr).state != IdleState {
		txt = "not idle state"
		return
	}

	(*userptr).state = EnteringRoomIdState
	txt = "enter room id"
	return
}

func joinRoomIdEnteredAction(userptr *User, userMessage string) (txt string) {
	roomId, err := strconv.ParseInt(userMessage, 10, 64)
	if err != nil {
		txt = "not a number"
		return
	}

	if !doesRoomExist(roomId) {
		txt = "room doesnt exist"
		return
	}

	(*userptr).state = InRoomState
	(*userptr).roomId = roomId
	txt = "entered room"
	return
}

func isUserOwner(userptr *User) bool {
	if (*userptr).roomId == (*userptr).id {
		return true
	}
	return false
}

func doesRoomExist(id int64) bool {
	for i := range rooms {
		if rooms[i].id == id {
			return true
		}
	}

	return false
}

func getUserById(userId int64) *User {
	for i := range users {
		if users[i].id == userId {
			return &users[i]
		}
	}
	
	users = append(users, User{userId, 0, IdleState})
	return &users[len(users)-1]
}

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_API_TOKEN"))
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		userId := update.FromChat().ID
		userptr := getUserById(userId)
		fmt.Println(*userptr)

		msg := tgbotapi.NewMessage(userId, "error")

		if update.CallbackQuery != nil {
			switch update.CallbackQuery.Data {
			case string(CreateRoomButtonData):
				msg.Text = createRoom(userptr)

			case string(JoinRoomButtonData):
				msg.Text = joinRoomButtonPressedAction(userptr)

			default:
				continue
			}

			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}
		} else if update.Message != nil {
			switch (*userptr).state {
			case IdleState:
				createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", string(CreateRoomButtonData))
				joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", string(JoinRoomButtonData))
				keyboard := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
				msg.ReplyMarkup = keyboard

			case EnteringRoomIdState:
				msg.Text = joinRoomIdEnteredAction(userptr, update.Message.Text)

			case InRoomState:
			}

		} else {
			continue
		}

		if _, err := bot.Send(msg); err != nil {
			panic(err)
		}
	}
}
