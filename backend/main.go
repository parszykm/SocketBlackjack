package main

import (
	// "blackjack/deck"

	"fmt"
	"net/http"

	"golang.org/x/net/websocket"
)

// WebSocket handler function
func wsHandler(ws *websocket.Conn) {
	fmt.Println("Client connected")

	// Send "Hello, World!" message to the client
	err := websocket.Message.Send(ws, "Hello, World!")
	if err != nil {
		fmt.Println("Error sending message:", err)
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

	// Set up a new WebSocket server at /ws endpoint
	http.Handle("/ws", websocket.Handler(wsHandler))

	// Start the HTTP server on port 8080
	fmt.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}

}
