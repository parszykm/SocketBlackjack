package blackjackserver

import (
	"blackjack/blackjack"
	"blackjack/deck"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"

	"golang.org/x/net/websocket"
)

type GameSession struct {
	Player *blackjack.Player
	Conn   *websocket.Conn
}

func StartServer(port string, game *blackjack.Game, room int) {
	roomStr := strconv.Itoa(room)
	http.Handle("/ws"+roomStr, websocket.Handler(func(ws *websocket.Conn) {
		wsHandler(ws, game)
	}))

	fmt.Printf("Starting server on %s:%s\n", "/ws"+roomStr, port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}

const (
	MessageTypeBind                = "Bind"
	MessageTypeInitialHandshake    = "InitialHandshake"
	MessageTypeStartGame           = "StartGame"
	MessageTypeEndGame             = "EndGame"
	MessageTypeHit                 = "Hit"
	MessageTypeStand               = "Stand"
	MessageTypeEndOfTurn           = "EndOfTurn"
	MessageTypeReconnect           = "Reconnect"
	MessageTypeReconnectResponse   = "ReconnectResponse"
	MessageTypeChangeStake         = "ChangeStake"
	MessageTypeChangeStakeResponse = "ChangeStakeResponse"
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MessageReconnectResponse struct {
	StoredId  int    `json:"storedId"`
	GivenId   int    `json:"givenId"`
	SessionId string `json:"sessionId"`
}

type MessageBind struct {
	Username  string  `json:"username"`
	Budget    float64 `json:"budget"`
	Stake     float64 `json:"stake"`
	OldId     int     `json:"oldId"`
	SessionId string  `json:"sessionId"`
}

var (
	mutex   sync.Mutex
	game    = blackjack.NewGame()
	counter = 0
)

func wsHandler(ws *websocket.Conn, game *blackjack.Game) {
	fmt.Println("Client connected")
	counter++
	fmt.Printf("Number of connections: %d\n", counter)

	newPlayer := &blackjack.Player{Username: "NewPlayer", Budget: 100}

	for {
		var msg WebSocketMessage

		if err := websocket.JSON.Receive(ws, &msg); err != nil {
			fmt.Printf("Client with ID %d disconnected...\n", newPlayer.Id)
			newPlayer.ActiveConnection = false
			mutex.Lock()

			counter--
			fmt.Printf("Number of connections: %d\n", counter)
			mutex.Unlock()
			break
		}

		switch msg.Type {
		case MessageTypeBind:
			fmt.Println("Received bind msg")
			dataMap, ok := msg.Data.(map[string]interface{})
			if !ok {
				fmt.Println("Error: Data is not in expected format")
				continue
			}
			bindMsg := MessageBind{
				Username:  dataMap["username"].(string),
				Budget:    dataMap["budget"].(float64),
				Stake:     dataMap["stake"].(float64),
				OldId:     int(dataMap["oldId"].(float64)),
				SessionId: dataMap["sessionId"].(string),
			}
			newPlayer.Username = bindMsg.Username
			newPlayer.Budget = bindMsg.Budget
			newPlayer.Id = bindMsg.OldId
			newPlayer.SessionId = bindMsg.SessionId
			fmt.Println(bindMsg)
			game.Bind(newPlayer, ws, bindMsg.Stake)
		case MessageTypeStartGame:
			fmt.Println("Received startGame msg")

			game.StartGame()
		case MessageTypeHit:
			fmt.Println("Received hit msg")
			newPlayer.Hit()
			newPlayer.ShowHand()

		case MessageTypeStand:
			for _, player := range game.Players {
				player.ShowHand()
			}
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
				StoredId:  int(dataMap["storedId"].(float64)),
				GivenId:   int(dataMap["givenId"].(float64)),
				SessionId: dataMap["sessionId"].(string),
			}

			fmt.Printf("Got ID of %d to reconnect\n", reconnectMsg.StoredId)

			found := false
			for _, player := range game.Players {
				fmt.Println(reconnectMsg.StoredId, player.Id, reconnectMsg.StoredId == player.Id)
				if player.Id == reconnectMsg.StoredId && player.SessionId == reconnectMsg.SessionId {
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
				fmt.Printf("Player with the ID of %d and sessionID %s has not been found. Binding previously given ID of %d...", reconnectMsg.StoredId, reconnectMsg.SessionId, reconnectMsg.GivenId)
				sendReconnectResponse(ws, reconnectMsg.GivenId)
			}
		case MessageTypeChangeStake:
			fmt.Println("Received CHAGE stake", msg.Data.(float64))
			if game.GameStage == blackjack.GameActive {
				res := WebSocketMessage{MessageTypeChangeStakeResponse, newPlayer.DefaultStake}
				sendMessage(ws, res)
			} else {
				newPlayer.DefaultStake = msg.Data.(float64)
				res := WebSocketMessage{MessageTypeChangeStakeResponse, newPlayer.DefaultStake}
				sendMessage(ws, res)
			}
		default:
			fmt.Println("Unknown message type:", msg.Type)
		}
	}

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

	_, err = ws.Write(jsonHand)
	if err != nil {
		fmt.Println("Error sending initial hand:", err)
	}
}

// func main() {

// 	http.Handle("/ws", websocket.Handler(wsHandler))

// 	fmt.Println("Starting server on :8080")
// 	if err := http.ListenAndServe(":8080", nil); err != nil {
// 		fmt.Println("Error starting server:", err)
// 	}

// }
