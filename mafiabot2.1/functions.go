package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getOrCreatePlayerById(playerId int64, playerName string) *Player {
	for i := range players {
		if players[i].id == playerId {
			return &players[i]
		}
	}
	players = append(players, Player{playerId, playerName, 0, IdleState, 0})
	return &players[len(players)-1]
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

func getKeyboard(state State) (buttons tgbotapi.InlineKeyboardMarkup) {
	switch state {
	case IdleState:
		createRoomButton := tgbotapi.NewInlineKeyboardButtonData("create", BtnCreate)
		joinRoomButton := tgbotapi.NewInlineKeyboardButtonData("join", BtnJoin)
		buttons = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(createRoomButton, joinRoomButton))

	case JoiningRoomState:
		leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("cancel", BtnLeave)
		buttons = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(leaveRoomButton))

	case InRoomIdleState:
		readyButton := tgbotapi.NewInlineKeyboardButtonData("ready", BtnReady)
		leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", BtnLeave)
		buttons = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(readyButton, leaveRoomButton))

	case InRoomReadyState:
		leaveRoomButton := tgbotapi.NewInlineKeyboardButtonData("leave", BtnLeave)
		buttons = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(leaveRoomButton))
	}

	return
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

	if getPlayerCountByOwner(ownerptr) < 4 {
		return false
	}

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
	if playerCount <= 0 {
		return []Role{}
	}

	roles = make([]Role, playerCount)

	mafiaCount := playerCount / 4
	if mafiaCount < 1 {
		mafiaCount = 1
	}

	for i := range roles {
		roles[i] = CivilianRole
	}

	for i := 0; i < mafiaCount; i++ {
		roles[i] = MafiaRole
	}

	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(roles), func(i, j int) {
		roles[i], roles[j] = roles[j], roles[i]
	})

	return
}

func assignRolesByOwner(ownerptr *Player) {
	roles := getRolesByOwner(ownerptr)
	playersInRoom := getPlayersByOwner(ownerptr)
	for i := range playersInRoom {
		playersInRoom[i].role = roles[i]
	}
}

func startGameByOwner(bot *tgbotapi.BotAPI, ownerptr *Player) {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			players[i].state = InGameState
		}
	}
	assignRolesByOwner(ownerptr)
	roleMessages := getRoleMessagesToAll(ownerptr)
	for i := range roleMessages {
		_, err := bot.Send(roleMessages[i])
		if err != nil {
			log.Print(err)
		}
	}
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

func changeStateToReady(playerptr *Player) {
	(*playerptr).state = InRoomReadyState

}

func sendMessageToOwner(bot *tgbotapi.BotAPI, ownerptr *Player, messageType string, playerptr *Player) {
	messageToOwner := tgbotapi.NewMessage(ownerptr.id, "")
	switch messageType {
	case "ready":
		messageToOwner.Text = fmt.Sprintf("%s ready", getReadyCountByOwner(ownerptr))
		_, err := bot.Send(messageToOwner)
		if err != nil {
			log.Print(err)
		}
	case "leave":
		messageToOwner.Text = fmt.Sprintf("%s left \n players in room: %d", playerptr.name, getPlayerCountByOwner(ownerptr))
		_, err := bot.Send(messageToOwner)
		if err != nil {
			log.Print(err)
		}
	case "join":
		messageToOwner.Text = fmt.Sprintf("%s joined \n players in room: %d", playerptr.name, getPlayerCountByOwner(ownerptr))
		_, err := bot.Send(messageToOwner)
		if err != nil {
			log.Print(err)
		}

	}
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
