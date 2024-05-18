package main

import (
	"blackjack/blackjack"
	"blackjack/deck"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type GameSession struct {
	Player *blackjack.Player
	Conn   *websocket.Conn
}

const (
	MessageTypeInitialHandshake  = "InitialHandshake"
	MessageTypeStartGame         = "StartGame"
	MessageTypeEndGame           = "EndGame"
	MessageTypeHit               = "Hit"
	MessageTypeStand             = "Stand"
	MessageTypeEndOfTurn         = "EndOfTurn"
	MessageTypeReconnect         = "Reconnect"
	MessageTypeReconnectResponse = "ReconnectResponse"
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MessageReconnectResponse struct {
	StoredId int `json:"storedId"`
	GivenId  int `json:"givenId"`
}

var (
	// Map to store WebSocket connections and associated players.
	// connections = make(map[*websocket.Conn]*blackjack.Player)
	mutex   sync.Mutex // Mutex for safe concurrent access to the map.
	game    = blackjack.NewGame()
	counter = 0
)

func wsHandler(ws *websocket.Conn) {
	fmt.Println("Client connected")
	counter++
	fmt.Printf("Number of connections: %d\n", counter)

	newPlayer := &blackjack.Player{Username: "NewPlayer", Budget: 100}

	game.Bind(newPlayer, ws, 10)

	// connections[ws] = player

	// sendInitialHandshake(ws)

	for {
		var msg WebSocketMessage

		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			fmt.Printf("Client with ID %d disconnected...\n", newPlayer.Id)
			newPlayer.ActiveConnection = false
			mutex.Lock()
			// delete(connections, ws)
			counter--
			fmt.Printf("Number of connections: %d\n", counter)
			mutex.Unlock()
			break
		}

		switch msg.Type {
		case MessageTypeStartGame:
			fmt.Println("Received startGame msg")

			game.StartGame()
		case MessageTypeHit:
			fmt.Println("Received hit msg")
			newPlayer.Hit()
			newPlayer.ShowHand()
			// if nextRound := player.ShowHand(); !nextRound {
			// 	break

		case MessageTypeStand:
			game.NextPlayer()
		case MessageTypeEndGame:
			game.EndGame()
			game.NewRound()
		case MessageTypeReconnect:
			dataMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				fmt.Println("Error: Data is not in expected format")
				continue
			}
			reconnectMsg := MessageReconnectResponse{
				StoredId: int(dataMap["storedId"].(float64)),
				GivenId:  int(dataMap["givenId"].(float64)),
			}

			fmt.Printf("Got ID of %d to reconnect\n", reconnectMsg.StoredId)
			found := false
			for _, player := range game.Players {
				if player.Id == reconnectMsg.StoredId {
					fmt.Printf("Reconnected %d...\n", reconnectMsg.StoredId)
					player.Reconnect(ws)
					sendReconnectResponse(ws, reconnectMsg.StoredId)
					found = true
					if reconnectMsg.StoredId != reconnectMsg.GivenId {
						game.DeletePlayer(reconnectMsg.GivenId)
					}
					newPlayer = player
					player.SendReconnectState()
					break
				}
			}
			if !found {
				fmt.Printf("Player with the ID of %d has not been found. Binding previously given ID of %d...", reconnectMsg.StoredId, reconnectMsg.GivenId)
				sendReconnectResponse(ws, reconnectMsg.GivenId)
			}

		default:
			fmt.Println("Unknown message type:", msg.Type)
		}
	}

	// game.StartGame()
	// sendInitialHand(ws, player)
}

func sendReconnectResponse(ws *websocket.Conn, id int) {
	msg := WebSocketMessage{
		Type: MessageTypeReconnectResponse,
		Data: id,
	}
	sendMessage(ws, msg)
}

func sendMessage(ws *websocket.Conn, msg WebSocketMessage) {
	if err := websocket.JSON.Send(ws, msg); err != nil {
		fmt.Println("Error sending message:", err)
	}
}

//	func sendDealersHand(ws *websocket.Conn, player *blackjack.Player) {
//		dealersHand := game
//	}
func sendInitialHand(ws *websocket.Conn, player *blackjack.Player) {
	myDeck := deck.New(deck.Deck(3), deck.Shuffle)

	hand := []deck.Card{}
	var tmpCard deck.Card
	rank := deck.Ace
	for i := 1; i <= 4; i++ {
		tmpCard = myDeck[rand.Int()%len(myDeck)]
		hand = append(hand, tmpCard)
		rank = rank + 1
	}

	jsonHand, err := json.Marshal(hand)
	if err != nil {
		fmt.Println("Error marshalling hand:", err)
		return
	}

	// Send the initial hand to the client.
	_, err = ws.Write(jsonHand)
	if err != nil {
		fmt.Println("Error sending initial hand:", err)
	}
}

func main() {
	// new_deck := deck.New(deck.Deck(3), deck.Shuffle)
	// deck.PrintDeck(new_deck)

	/*
		game := blackjack.NewGame()
		player1 := &blackjack.Player{Username: "majkel", Budget: 20}
		player2 := &blackjack.Player{Username: "ziomo2", Budget: 3000}
		game.Bind(player1, 10)
		game.Bind(player2, 1500)
		game.StartGame()
	*/
	http.Handle("/ws", websocket.Handler(wsHandler))

	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}

}
