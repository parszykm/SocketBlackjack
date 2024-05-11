package blackjack

import (
	deck "blackjack/deck"
	"bufio"
	"fmt"
	"os"
)

// type IPlayer interface {
// 	EvalHand()
// 	Hit() (deck.Card, error)
// 	ShowHand()
// }

type Game struct {
	Deck    []deck.Card
	Players []*Player
	Dealer  *Dealer
}

type Player struct {
	Username string
	Budget   float64
	Stake    float64
	Game     *Game
	Hand     []deck.Card
	Count    int
}

type Dealer struct {
	GameStage bool
	Player
}

func NewGame() *Game {
	game := new(Game)
	game.Deck = deck.New(deck.Deck(1), deck.Shuffle)
	game.Dealer = new(Dealer)
	game.Dealer.GameStage = false
	game.Dealer.Game = game

	game.Dealer.Hit()
	game.Dealer.Hit()

	game.Dealer.ShowHand()

	return game
}

func (g *Game) StartGame() {
	for i := 0; i < 2; i++ {
		for _, player := range g.Players {
			player.Hit()
		}
	}
	for _, player := range g.Players {
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
	g.EndGame()
}
func (g *Game) EndGame() error {
	g.Dealer.GameStage = true
	g.Dealer.ShowHand()

	for g.Dealer.EvalHand(); g.Dealer.EvalHand() < 17; g.Dealer.Hit() {
		fmt.Println("Dealer's hands is less then 17. Drawing...")
		g.Dealer.ShowHand()
		if g.Dealer.EvalHand() > 21 {
			fmt.Println("Everybody wins!")
			for _, player := range g.Players {
				player.Budget += player.Stake * 2
			}
			return nil
		}
	}

	dCount := g.Dealer.EvalHand()
	var currentWin float64 = 0.0
	for _, player := range g.Players {
		currentWin = 0.0
		if player.EvalHand() == 21 {
			currentWin = player.Stake * 2.5
			player.Budget += currentWin
			fmt.Printf("%s has BLACKJACK and wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() < 21 && player.EvalHand() > dCount {
			currentWin = player.Stake * 2
			player.Budget += currentWin
			player.Stake = 0
			fmt.Printf("%s Wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() == dCount {
			currentWin = player.Stake
			player.Budget += currentWin
			player.Stake = 0
			fmt.Printf("Push. %s ties with a dealer. Refunding... %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else {
			player.Stake = 0
			fmt.Printf("%s loses!\nCurrent budget: %f\n", player.Username, player.Budget)
		}

	}

	return nil
}

func (g *Game) Bind(p *Player, stake float64) error {
	if p.Budget < stake {
		return fmt.Errorf("Not enough credits in player's budget...")
	}
	p.Budget -= stake
	p.Stake = stake
	p.Game = g
	g.Players = append(g.Players, p)
	return nil
}

// func (p *Player) Bind(game *Game) {
// 	p.Game = game
// 	game.Players = append(game.Players, p)
// }

func (p *Player) Hit() (deck.Card, error) {

	if len(p.Game.Deck) == 0 {
		return deck.Card{}, fmt.Errorf("Deck is empty")
	}
	pop := p.Game.Deck[len(p.Game.Deck)-1]
	p.Game.Deck = p.Game.Deck[:len(p.Game.Deck)-1]
	p.Hand = append(p.Hand, pop)
	return pop, nil
}

func (p *Player) EvalHand() int {
	p.Count = 0
	for _, card := range p.Hand {
		p.Count += card.GetValue()
	}
	return p.Count
}

func (p *Player) ShowHand() bool {
	fmt.Printf("%s hand:\n", p.Username)
	for _, card := range p.Hand {
		fmt.Printf("%s, ", card)
	}
	fmt.Println()
	count := p.EvalHand()
	if count == 21 {
		// p.Budget += p.Stake * 2.5
		fmt.Printf("Count: %d\n-----------\n", p.Count)
		fmt.Println("You have BLACKJACK. Congrats...")
		return false
	} else if count > 21 {
		// p.Stake = 0
		fmt.Printf("Count: %d\n-----------\n", p.Count)
		fmt.Println("Your hand exceeded the 21 score. You lost...")
		return false
	}
	fmt.Printf("Count: %d\n-----------\n", p.Count)
	return true
}

func (d *Dealer) ShowHand() {
	fmt.Printf("Dealer's hand:\n")
	for i, card := range d.Hand {
		if i == 0 || d.GameStage {
			fmt.Printf("%s, ", card)
		} else {
			fmt.Printf("***HIDDEN***")
		}
	}
	if d.GameStage {
		d.EvalHand()
		fmt.Println(d.Count, "\n-------")
	} else {
		fmt.Println("\n-------")
	}
}
