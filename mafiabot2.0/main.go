package main

import (
	"fmt"
	"math/rand"
	"strconv"

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

type State uint

const (
	IdleState State = iota
	JoiningRoomState
	InRoomIdleState
	InRoomReadyState
	InGameState
)

type Role uint

const (
	CivilianRole Role = iota
	MafiaRole
)

func createPlayer(playerId int64, playerName string) *Player {
	for i := range players {
		if playerId == players[i].id {
			return nil
		}
	}
	players = append(players, Player{playerId, playerName, 0, IdleState, 0})
	return &players[len(players)-1]
}

func getPlayerByID(playerId int64, playerName string) *Player {
	for i := range players {
		if players[i].id == playerId {
			return &players[i]
		}
	}
	playerptr := createPlayer(playerId, playerName)
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
	return (*playerptr).id == (*playerptr).room
}

func getOwnerByPLayer(playerptr *Player) *Player {
	for i := range players {
		if players[i].id == (*playerptr).room {
			return &players[i]
		}
	}
	return nil
}

func getIdleButtons() tgbotapi.InlineKeyboardMarkup {
	createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", "createRoom")
	joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", "joinRoom")
	buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))
	return buttons
}

func getInRoomIdleButtons() tgbotapi.InlineKeyboardMarkup {
	readyButton := tgbotapi.NewInlineKeyboardButtonData("ready", "ready")
	leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", "leaveRoom")
	buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(readyButton, leaveRoomButton))
	return buttons
}

func getInRoomReadyButtons() tgbotapi.InlineKeyboardMarkup {
	leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", "leaveRoom")
	buttons := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(leaveRoomButton))
	return buttons
}

func getPlayerCountByOwner(ownerptr *Player) (playerCount int) {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			playerCount++
		}
	}
	return
}

func getReadyCountByOwner(ownerptr *Player) string {
	var allPlayers int
	var readyPlayers int
	for i := range players {
		if players[i].room == (*ownerptr).id {
			allPlayers++
			if players[i].state == InRoomReadyState {
				readyPlayers++
			}
		}
	}
	return fmt.Sprintf("%d/%d", readyPlayers, allPlayers)
}

func canGameStart(ownerptr *Player) bool {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			if players[i].state != InRoomReadyState {
				return false
			}
		}
	}
	return true
}

func getRolesByOwner(ownerptr *Player) (roles []Role) {
	playerCount := getPlayerCountByOwner(ownerptr)
	roles = make([]Role, playerCount)

	if playerCount <= 0 {
		return []Role{}
	}

	for i := range roles {
		roles[i] = CivilianRole
	}

	roles[rand.Intn(playerCount)] = MafiaRole
	return
}

func assignRolesByOwner(ownerptr *Player) {
	roles := getRolesByOwner(ownerptr)
	playersInRoom := getPlayersByOwner(ownerptr)
	for i := range playersInRoom {
		playersInRoom[i].role = roles[i]
	}
}

func startGameByOwner(ownerptr *Player) {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			players[i].state = InGameState
		}
	}
	assignRolesByOwner(ownerptr)
}

func getPlayersByOwner(ownerptr *Player) (playersInRoom []*Player) {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			playersInRoom = append(playersInRoom, &players[i])
		}
	}
	return
}

func getSameMessagesToAll(ownerptr *Player, messageText string) (messages []tgbotapi.MessageConfig) {
	playersInRoom := getPlayersByOwner(ownerptr)
	for i := range playersInRoom {
		messages = append(messages, tgbotapi.NewMessage(playersInRoom[i].id, messageText))
	}
	return messages
}

func getRoleMessagesToAll(ownerptr *Player) (messages []tgbotapi.MessageConfig) {
	playersInRoom := getPlayersByOwner(ownerptr)

	for _, p := range playersInRoom {
		var text string
		if p.role == MafiaRole {
			text = "game started \n your role: mafia"
		} else {
			text = "game started \n your role: civilian"
		}
		messages = append(messages, tgbotapi.NewMessage(p.id, text))
	}
	return messages
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
		var name string

		if update.Message != nil {
			chatID = update.Message.Chat.ID
			userID = update.Message.From.ID
			name = update.Message.From.FirstName
		} else if update.CallbackQuery != nil {
			chatID = update.CallbackQuery.Message.Chat.ID
			userID = update.CallbackQuery.From.ID
			name = update.CallbackQuery.From.FirstName
			bot.Request(tgbotapi.NewCallback(update.CallbackQuery.ID, ""))
		} else {
			continue
		}

		playerptr := getPlayerByID(userID, name)
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
						response.ReplyMarkup = getInRoomIdleButtons()
					} else {
						response.Text = "error creating room"
					}

				case "joinRoom":
					if (*playerptr).room == 0 {
						changeStateToJoining(playerptr)
						response.Text = "enter room id"
					} else {
						response.Text = "already in room"
						response.ReplyMarkup = getIdleButtons()
					}
				}

			case JoiningRoomState:
				changeStateToIdle(playerptr)
				response.Text = "cancelled join"
				response.ReplyMarkup = getIdleButtons()

			case InRoomIdleState:
				switch playerQuery {
				case "ready":
					changeStateToReady(playerptr)

					ownerptr := getOwnerByPLayer(playerptr)

					if ownerptr != nil {
						ownerId := ownerptr.id
						readyCount := getReadyCountByOwner(ownerptr)
						responseToOwner := tgbotapi.NewMessage(ownerId, fmt.Sprintf("ready to play: %s", readyCount))
						if _, err := bot.Send(responseToOwner); err != nil {
							fmt.Println("error sending message", err)
						}
						if canGameStart(ownerptr) {
							startGameByOwner(ownerptr)
							messagesToAll := getRoleMessagesToAll(ownerptr)
							for i := range messagesToAll {
								if _, err := bot.Send(messagesToAll[i]); err != nil {
									fmt.Println("error sending message", err)
								}
							}
							break
						}
					}

					response.Text = "wait for game to start"
					response.ReplyMarkup = getInRoomReadyButtons()

				case "leaveRoom":
					if ownerptr := getOwnerByPLayer(playerptr); ownerptr != nil {
						ownerId := ownerptr.id
						leaveRoom(playerptr)
						playerCount := getPlayerCountByOwner(ownerptr)
						responseToOwner := tgbotapi.NewMessage(ownerId, fmt.Sprintf("user %d left, player count: %d", (*playerptr).id, playerCount))
						if _, err := bot.Send(responseToOwner); err != nil {
							fmt.Println("error sending message", err)
						}
					}

					response.Text = "you left the room"
					response.ReplyMarkup = getIdleButtons()
				}
			case InRoomReadyState:
				if playerQuery == "leaveRoom" {
					if ownerptr := getOwnerByPLayer(playerptr); ownerptr != nil {
						ownerId := ownerptr.id
						leaveRoom(playerptr)
						playerCount := getPlayerCountByOwner(ownerptr)
						responseToOwner := tgbotapi.NewMessage(ownerId, fmt.Sprintf("user %d left, player count: %d", (*playerptr).id, playerCount))
						if _, err := bot.Send(responseToOwner); err != nil {
							fmt.Println("error sending message", err)
						}
					}
					leaveRoom(playerptr)
					response.Text = "you left the room"
					response.ReplyMarkup = getIdleButtons()
				} else {
					response.Text = "wait for game to start"
				}
			}

		} else if update.Message != nil {
			playerMessage := update.Message.Text
			switch (*playerptr).state {
			case JoiningRoomState:
				if addPlayerToRoom(playerptr, playerMessage) {
					if ownerptr := getOwnerByPLayer(playerptr); ownerptr != nil {
						ownerId := ownerptr.id
						playerCount := getPlayerCountByOwner(ownerptr)
						responseToOwner := tgbotapi.NewMessage(ownerId, fmt.Sprintf("user %d joined, player count: %d", (*playerptr).id, playerCount))
						if _, err := bot.Send(responseToOwner); err != nil {
							fmt.Println("error sending message", err)
						}
					}

					response.Text = "room entered"
					response.ReplyMarkup = getInRoomIdleButtons()
				} else {
					changeStateToIdle(playerptr)
					response.Text = "error invalid id"
					response.ReplyMarkup = getIdleButtons()
				}
			case IdleState:
				response.Text = "create or join room"
				response.ReplyMarkup = getIdleButtons()
			default:
				response.Text = "error"
			}
		}

		if response.Text != "" {
			if _, err := bot.Send(response); err != nil {
				fmt.Println("error sending message", err)
			}
		}
	}
}
