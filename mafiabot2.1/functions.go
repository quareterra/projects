package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// GAME SETUP //

func getRoomKeyboard(state State) (buttons tgbotapi.InlineKeyboardMarkup) {
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

func getOrCreatePlayerById(playerId int64, playerName string) *Player {
	for i := range players {
		if players[i].id == playerId {
			return &players[i]
		}
	}
	players = append(players, Player{playerId, playerName, 0, IdleState, 0, false, false})
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

// IN GAME //

func inGameAction(bot *tgbotapi.BotAPI, playerptr *Player, callback *tgbotapi.CallbackQuery) {
	if playerptr.didAction == true || playerptr.alive == false {
		return
	}

	ownerptr := getOwnerByPLayer(playerptr)
	playersInRoom := getPlayersByOwner(ownerptr)
	var targetPlayer *Player
	for i := range playersInRoom {
		if playersInRoom[i].name == callback.Data {
			targetPlayer = playersInRoom[i]
		}
	}

	if targetPlayer == nil {
		return
	}

	roomptr := getRoomByPlayer(playerptr)
	isNight := roomptr.turn%2 == 1

	if isNight {
		switch playerptr.role {
		case MafiaRole:
			var playerSlice []*Player
			playerSlice = append(playerSlice, targetPlayer)
			roomptr.actedUpon = append(roomptr.actedUpon, targetPlayer)
			kill(playerSlice)
			playerptr.didAction = true
		}
	} else {
		playerptr.didAction = true
		bot.Send(tgbotapi.NewMessage(playerptr.id, fmt.Sprintf("You voted for %s", targetPlayer.name)))
		roomptr.actedUpon = append(roomptr.actedUpon, targetPlayer)
	}

	if canTurnEnd(ownerptr) {
		nextTurn(bot, ownerptr)
	}
}

func playerSliceToStr(playerSLice []*Player) (playerStr string) {
	for i := range playerSLice {
		playerStr = fmt.Sprint(playerStr, playerSLice[i].name)
	}
	return
}

func calculateResults(bot *tgbotapi.BotAPI, ownerptr *Player) {
	roomptr := getRoomByPlayer(ownerptr)
	playersInRoom := getPlayersByOwner(ownerptr)

	isNight := roomptr.turn%2 == 1

	if isNight {
		diedThisNight := roomptr.actedUpon
		sendMessageToAll(bot, playersInRoom, fmt.Sprintf("died this night %s", playerSliceToStr(diedThisNight)), nil)
	} else {
		votedOut := roomptr.actedUpon
		sendMessageToAll(bot, votedOut, fmt.Sprintf("votedOut: %d", playerSliceToStr(votedOut)), nil)
	}
}

func canTurnEnd(ownerptr *Player) bool {
	playersInRoom := getPlayersByOwner(ownerptr)
	for i := range playersInRoom {
		if playersInRoom[i].alive && playersInRoom[i].didAction == false {
			return false
		}
	}
	return true
}

func killVoted(votes []*Player, ownerptr *Player) {
	repeats := make(map[*Player]int)
	for _, i := range votes {
		repeats[i]++
	}
	maxRepeats := 0
	var mostRepeated *Player
	for i, j := range repeats {
		if j > maxRepeats {
			maxRepeats = j
			mostRepeated = i
		}
	}
	if mostRepeated != nil {
		playersToKill := []*Player{mostRepeated}
		roomptr := getRoomByPlayer(ownerptr)
		roomptr.actedUpon = playersToKill
		kill(playersToKill)
	}
}

func nextTurn(bot *tgbotapi.BotAPI, ownerptr *Player) {
	room := getRoomByPlayer(ownerptr)
	if room.turn%2 == 0 {
		killVoted(room.actedUpon, ownerptr)
	}
	calculateResults(bot, ownerptr)
	room.turn++
	room.actedUpon = []*Player{}

	playersInRoom := getPlayersByOwner(ownerptr)
	for i := range playersInRoom {
		playersInRoom[i].didAction = false
	}

	if room.turn%2 == 1 {
		startNight(bot, ownerptr)
	} else {
		startDay(bot, ownerptr)
	}
}

func getAlive(playersInRoom []*Player) (alivePlayers []*Player) {
	for i := range playersInRoom {
		alivePlayers = append(alivePlayers, playersInRoom[i])
	}
	return
}

func startDay(bot *tgbotapi.BotAPI, ownerptr *Player) {
	playersInRoom := getPlayersByOwner(ownerptr)
	alivePlayers := getAlive(playersInRoom)
	actionKeyboard := getGameKeyboard(alivePlayers)
	sendMessageToAll(bot, playersInRoom, "choose who will die:", &actionKeyboard)
}

func startNight(bot *tgbotapi.BotAPI, ownerptr *Player) {
	playersInRoom := getPlayersByOwner(ownerptr)
	alivePlayers := getAlive(playersInRoom)
	var mafiaPlayers, civilianPlayers []*Player

	for i := range alivePlayers {
		if alivePlayers[i].role == MafiaRole {
			mafiaPlayers = append(mafiaPlayers, alivePlayers[i])
			alivePlayers[i].didAction = false
		} else if alivePlayers[i].role == CivilianRole {
			alivePlayers[i].didAction = true
			civilianPlayers = append(civilianPlayers, alivePlayers[i])
		}
	}

	killKeyboard := getGameKeyboard(civilianPlayers)

	sendMessageToAll(bot, mafiaPlayers, "choose who will die:", &killKeyboard)
	sendMessageToAll(bot, civilianPlayers, "survive the night (just wait)", nil)
}

func getGameKeyboard(playerSlice []*Player) (buttons tgbotapi.InlineKeyboardMarkup) {
	var buttonSlice []tgbotapi.InlineKeyboardButton
	for i := range playerSlice {
		buttonSlice = append(buttonSlice, tgbotapi.NewInlineKeyboardButtonData(playerSlice[i].name, playerSlice[i].name))
	}
	buttons = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttonSlice...))
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

	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	r.Shuffle(len(roles), func(i, j int) {
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

func revive(playerSlice []*Player) {
	for i := range playerSlice {
		playerSlice[i].alive = true
	}
}

func kill(playerSlice []*Player) {
	for i := range playerSlice {
		playerSlice[i].alive = false
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
	for i := range rooms {
		if rooms[i].id == ownerptr.id {
			rooms[i].turn = 1
		}
	}
	playersInRoom := getPlayersByOwner(ownerptr)
	revive(playersInRoom)
	startNight(bot, ownerptr)
}

func getPlayersByOwner(ownerptr *Player) (playersInRoom []*Player) {
	for i := range players {
		if players[i].room == (*ownerptr).id {
			playersInRoom = append(playersInRoom, &players[i])
		}
	}
	return
}

func sendMessageToAll(bot *tgbotapi.BotAPI, playerSlice []*Player, messageText string, keyboardptr *tgbotapi.InlineKeyboardMarkup) {
	var messages []tgbotapi.MessageConfig
	for i := range playerSlice {
		message := tgbotapi.NewMessage(playerSlice[i].id, messageText)
		if keyboardptr != nil {
			message.ReplyMarkup = *keyboardptr
		}
		messages = append(messages, message)
	}
	for i := range messages {
		bot.Send(messages[i])
	}
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
		rooms = append(rooms, Room{(*playerptr).id, 0, nil})
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
