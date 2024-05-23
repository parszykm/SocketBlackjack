package main

import (
	"blackjack/blackjack"
	"blackjack/blackjackserver"
)

func main() {
	// blackjackserver.StartServer("8080")
	ports := []string{"8080", "8081", "8082", "8083", "8084"}

	for index, port := range ports {
		game := blackjack.NewGame()
		go blackjackserver.StartServer(port, game, index)
	}

	// Block forever
	select {}
}
