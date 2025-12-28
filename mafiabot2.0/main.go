package main

import (
	"fmt"

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
	JoiningRoomState
	InRoomState
)

func createPlayer(playerId int64) *Player {
	for i := range players {
		if playerId == players[i].id {
			return nil
		}
	}
	players = append(players, Player{playerId, 0})
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

func createRoom(playerptr *Player) *Room {
	if (*playerptr).room == 0 {
		rooms = append(rooms, Room{(*playerptr).id})
		(*playerptr).room = rooms[len(rooms)-1].id
		return pl
	}
	return nil
}

func main() {
	/* setup bot */
	bot, err := tgbotapi.NewBotAPI("7401254673:AAGR-g_Ur41t9d1DgjRzK7uvxhmr7CSCBVs")
	if err != nil {
		panic(err)
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := bot.GetUpdatesChan(updateConfig)
	for update := range updates {
		playerId := update.Message.From.ID
		playerMessage := update.Message.Text
		playerptr := getPlayerByID(playerId)

		response := tgbotapi.NewMessage(playerId, "error")

		switch playerMessage {
		case "/createRoom":
			roomptr := createRoom(playerptr)
			if roomptr != nil {
				response.Text = fmt.Sprintf("Created room %d", (*roomptr).id)
			} else {
				response.Text = "Error"
			}
		}
		if _, err := bot.Send(response); err != nil {
			panic(err)
		}
	}
}
