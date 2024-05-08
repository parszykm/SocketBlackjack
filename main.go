package main

import (
	// "blackjack/deck"
	"blackjack/blackjack"
	"bufio"
	"fmt"
	"os"
)

func main() {
	// new_deck := deck.New(deck.Deck(3), deck.Shuffle)
	// deck.PrintDeck(new_deck)

	game := blackjack.NewGame()
	player1 := &blackjack.Player{Username: "majkel", Budget: 20}
	player2 := &blackjack.Player{Username: "ziomo2", Budget: 3000}
	game.Bind(player1, 10)
	game.Bind(player2, 1500)
	game.StartGame()
	for _, player := range game.Players {
		for {
			player.ShowHand()
			fmt.Println("Press either H for hit or S for stay")
			reader := bufio.NewReader(os.Stdin)
			choice, _ := reader.ReadByte()
			if choice == 'H' {
				player.Hit()
				if nextRound := player.ShowHand(); !nextRound {
					break
				}

			} else if choice == 'S' {
				break
			}
		}
	}
	game.EndGame()

}
