package main

import (
	"fmt"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Room struct {
	id int64
}

type Player struct {
	id    int64
	room  int64
	state State
}

var (
	rooms   []Room
	players []Player
)

type State uint

const (
	IdleState State = iota
	JoiningRoomState
	InRoomIdleState
	InRoomReadyState
)

func createPlayer(playerId int64) *Player {
	for i := range players {
		if playerId == players[i].id {
			return nil
		}
	}
	players = append(players, Player{playerId, 0, IdleState})
	return &players[len(players)-1]
}

func getPlayerByID(playerId int64) *Player {
	for i := range players {
		if players[i].id == playerId {
			return &players[i]
		}
	}
	playerptr := createPlayer(playerId)
	return playerptr
}

func getRoomByPlayer(playerptr *Player) *Room {
	for i := range rooms {
		if rooms[i].id == (*playerptr).room {
			return &rooms[i]
		}
	}
	return nil
}

func isOwner(playerptr *Player) bool {
	if (*playerptr).id == (*playerptr).room {
		return true
	}
	return false
}

func createRoom(playerptr *Player) *Room {
	if (*playerptr).room == 0 {
		rooms = append(rooms, Room{(*playerptr).id})
		(*playerptr).room = rooms[len(rooms)-1].id
		(*playerptr).state = InRoomIdleState
		return &rooms[len(rooms)-1]
	}
	return nil
}

func leaveRoom(playerptr *Player) {
	if isOwner(playerptr) {
		for i := range rooms {
			if rooms[i].id == (*playerptr).room {
				for j := range players {
					if players[j].room == rooms[i].id {
						players[j].room = 0
						players[j].state = IdleState
					}
				}
				rooms = append(rooms[:i], rooms[i+1:]...)
				return
			}
		}
	}
	(*playerptr).room = 0
	(*playerptr).state = IdleState
}

func changeStateToIdle(playerptr *Player) {
	(*playerptr).state = IdleState
}

func changeStateToJoining(playerptr *Player) {
	(*playerptr).state = JoiningRoomState
}

func changeStateToReady(playerptr *Player) {
	(*playerptr).state = InRoomReadyState
}

func addPlayerToRoom(playerptr *Player, playerMessage string) bool {
	roomId, err := strconv.ParseInt(playerMessage, 10, 64)
	if err != nil {
		return false
	}
	for i := range rooms {
		if rooms[i].id == roomId {
			(*playerptr).room = rooms[i].id
			(*playerptr).state = InRoomIdleState
			return true
		}
	}
	return false
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
		fmt.Println(rooms)
		fmt.Println(players)

		var chatID int64
		var userID int64

		if update.Message != nil {
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
		} else if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
			userID = update.CallbackQuery.From.ID
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		} else {
			continue
		}

		playerptr := getPlayerByID(userID)
		response := tgbotapi.NewMessage(chatID, "")

		if update.CallbackQuery != nil {
			playerQuery := update.CallbackQuery.Data
			switch (*playerptr).state {
			case IdleState:
				switch playerQuery {
				case "createRoom":
					roomptr := createRoom(playerptr)

					if roomptr != nil {
						response.Text = fmt.Sprintf("Created room %d", (*roomptr).id)
						readyButton := tgbotapi.NewInlineKeyboardButtonData("ready", "ready")
						leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", "leaveRoom")
						buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(readyButton, leaveRoomButton))
						response.ReplyMarkup = buttons
					} else {
						response.Text = "Error creating room"
					}

				case "joinRoom":
					if (*playerptr).room == 0 {
						changeStateToJoining(playerptr)
						response.Text = "enter room id"
					} else {
						response.Text = "already in room"
						createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
						joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
						buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
						response.ReplyMarkup = buttons
					}
				}

			case JoiningRoomState:
				changeStateToIdle(playerptr)
				response.Text = "cancelled"
				createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
				joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
				buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
				response.ReplyMarkup = buttons

			case InRoomIdleState:
				switch playerQuery {
				case "ready":
					changeStateToReady(playerptr)
					response.Text = "wait for game to start"
					leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", "leaveRoom")
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(leaveRoomButton))
					response.ReplyMarkup = buttons

				case "leaveRoom":
					leaveRoom(playerptr)
					response.Text = "you left the room"
					createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
					joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
					response.ReplyMarkup = buttons
				}
			case InRoomReadyState:
				if playerQuery == "leaveRoom" {
					leaveRoom(playerptr)
					response.Text = "you left the room"
					createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
					joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
					response.ReplyMarkup = buttons
				} else {
					response.Text = "wait for game to start"
				}
			}

		} else if update.Message != nil {
			playerMessage := update.Message.Text
			switch (*playerptr).state {
			case JoiningRoomState:
				if addPlayerToRoom(playerptr, playerMessage) {
					response.Text = "room entered"
					readyButton := tgbotapi.NewInlineKeyboardButtonData("ready", "ready")
					leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", "leaveRoom")
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(readyButton, leaveRoomButton))
					response.ReplyMarkup = buttons
				} else {
					changeStateToIdle(playerptr)
					response.Text = "error invalid id"
					createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
					joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
					buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
					response.ReplyMarkup = buttons
				}
			default:
				response.Text = "error"
				createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
				joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
				buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
				response.ReplyMarkup = buttons
			}
		}

		if response.Text != "" {
			if _, err := bot.Send(response); err != nil {
				fmt.Println("error sending message", err)
			}
		}
	}
}

