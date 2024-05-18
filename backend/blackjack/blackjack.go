package blackjack

import (
	deck "blackjack/deck"
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

type GameStage uint8

const (
	PreGame GameStage = iota
	GameActive
	PostGame
)
const (
	MessageTypeSendHand          = "SendHand"
	MessageTypeDealerInitHand    = "DealerInitHand"
	MessageTypeDealerFinalHand   = "DealerFinalHand"
	MessageTypeStartGameResponse = "StartGameResponse"
	MessageTypeGameResult        = "GameResult"
	MessageTypeInitialHandshake  = "InitialHandshake"
	MessageTypeStartGame         = "StartGame"
	MessageTypeOtherHands        = "OtherHands"
	MessageTypeYourTurn          = "YourTurn"
	MessageTypeReconnectState    = "ReconnectState"
)

type WebSocketMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type InitialHandshakeMessage struct {
	Id  int    `json:"id"`
	Msg string `json:"msg"`
}
type HandMessage struct {
	Hand  []deck.Card `json:"hand"`
	Stage bool        `json:"stage"`
	Count int         `json:"count"`
}

type ResultMessage struct {
	Count  int     `json:"count"`
	Refund float64 `json:"refund"`
	Budget float64 `json:"budget"`
}

type OtherHandStruct struct {
	Id        int         `json:"id"`
	Hand      []deck.Card `json:"hand"`
	Timestamp time.Time   `json:"timestamp"`
}

type ReconnectStateMessage struct {
	DealerHand []deck.Card       `json:"dealerHand"`
	Hand       []deck.Card       `json:"hand"`
	OtherHands []OtherHandStruct `json:"otherHands"`
	Turn       bool              `json:"turn"`
	GameStage  GameStage         `json:"gameStage"`
	Count      int               `json:"count"`
}

type Game struct {
	mutex            sync.Mutex
	PlayerIdCounters int
	Deck             []deck.Card
	Players          []*Player
	Dealer           *Dealer
	CurrentTurn      int
	GameStage        GameStage
}

type Player struct {
	Id               int
	Username         string
	Budget           float64
	Stake            float64
	Game             *Game
	Hand             []deck.Card
	Count            int
	conn             *websocket.Conn
	ActiveConnection bool
}

type Dealer struct {
	Player
}

func sendMessage(ws *websocket.Conn, msg WebSocketMessage) {
	if err := websocket.JSON.Send(ws, msg); err != nil {
		fmt.Println("Error sending message:", err)
	}
}

func NewGame() *Game {
	game := new(Game)
	game.PlayerIdCounters = 0
	game.Deck = deck.New(deck.Deck(1), deck.Shuffle)
	game.Dealer = new(Dealer)
	game.GameStage = PreGame
	game.Dealer.Game = game
	game.CurrentTurn = 0
	return game
}

func (g *Game) FilterPlayers() {
	var ActivePlayers []*Player
	fmt.Println(g.Players)
	for _, player := range g.Players {
		if player.ActiveConnection {
			ActivePlayers = append(ActivePlayers, player)
		}
	}
	g.Players = ActivePlayers
	fmt.Println(g.Players)
}
func (g *Game) NewRound() {
	g.FilterPlayers()
	fmt.Println(g.Dealer.Hand)
	g.CurrentTurn = 0
	g.Deck = deck.New(deck.Deck(1), deck.Shuffle)
	g.Dealer = new(Dealer)
	g.Dealer.Game = g
	g.GameStage = PreGame

	for _, player := range g.Players {
		player.Hand = nil
	}

	fmt.Println(g.Dealer.Hand)
}
func (g *Game) StartGame() {
	g.FilterPlayers()
	g.GameStage = GameActive
	fmt.Println(g.Dealer.Hand)
	msg := WebSocketMessage{MessageTypeStartGame, "StartGame"}
	for _, player := range g.Players {
		sendMessage(player.conn, msg)
	}
	g.Dealer.Hit()
	g.Dealer.Hit()

	g.Dealer.ShowHand()
	for i := 0; i < 2; i++ {
		for _, player := range g.Players {
			player.Hit()
		}
	}
	for _, player := range g.Players {
		player.ShowHand()
	}
	g.NextPlayer()
}

func (g *Game) DeletePlayer(id int) {
	for _, player := range g.Players {
		if player.Id == id {
			player.ActiveConnection = false
			break
		}
	}
}
func (p *Player) Reconnect(ws *websocket.Conn) {
	p.conn = ws
	p.ActiveConnection = true
}
func (g *Game) NextPlayer() {
	if g.CurrentTurn == len(g.Players) {
		g.EndGame()
		g.NewRound()
		return
	}
	player := g.Players[g.CurrentTurn]
	if !player.ActiveConnection {
		g.CurrentTurn++
		g.NextPlayer()
		return
	}
	msg := WebSocketMessage{MessageTypeYourTurn, 1}
	sendMessage(player.conn, msg)
	g.CurrentTurn++
}
func (g *Game) EndGame() error {
	g.GameStage = PostGame
	g.Dealer.ShowHand()

	for g.Dealer.EvalHand() < 17 {
		g.Dealer.Hit()
		fmt.Println("Dealer's hands is less then 17. Drawing...")
		g.Dealer.ShowHand()
		if g.Dealer.EvalHand() > 21 {
			fmt.Println("Everybody wins!")

			for _, player := range g.Players {
				if !player.ActiveConnection {
					continue
				}
				if player.EvalHand() <= 21 {
					player.Budget += player.Stake * 2
					message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: player.Stake * 2, Budget: player.Budget}}
					sendMessage(player.conn, message)
				}
			}
			return nil
		}
	}

	dCount := g.Dealer.EvalHand()
	var currentWin float64 = 0.0
	for _, player := range g.Players {
		if !player.ActiveConnection {
			continue
		}
		currentWin = 0.0
		if player.EvalHand() == 21 {
			currentWin = player.Stake * 2.5
			player.Budget += currentWin
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s has BLACKJACK and wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() < 21 && player.EvalHand() > dCount {
			currentWin = player.Stake * 2
			player.Budget += currentWin
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s Wins %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else if player.EvalHand() == dCount {
			currentWin = player.Stake
			player.Budget += currentWin
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: currentWin, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("Push. %s ties with a dealer. Refunding... %f!\nCurrent budget: %f\n", player.Username, currentWin, player.Budget)
		} else {
			player.Stake = 0
			message := WebSocketMessage{MessageTypeGameResult, ResultMessage{Count: player.EvalHand(), Refund: 0, Budget: player.Budget}}
			sendMessage(player.conn, message)
			fmt.Printf("%s loses!\nCurrent budget: %f\n", player.Username, player.Budget)
		}

	}

	return nil
}

func (g *Game) Bind(p *Player, ws *websocket.Conn, stake float64) (int, error) {
	if p.Budget < stake {
		return -1, fmt.Errorf("Not enough credits in player's budget...")
	}
	g.mutex.Lock()
	defer g.mutex.Unlock()

	p.Id = g.PlayerIdCounters
	g.PlayerIdCounters++
	p.Budget -= stake
	p.Stake = stake
	p.Game = g
	p.conn = ws
	p.ActiveConnection = true
	g.Players = append(g.Players, p)
	fmt.Printf("New player hes been gived ID of  %d\n", p.Id)
	msg := WebSocketMessage{MessageTypeInitialHandshake, InitialHandshakeMessage{p.Id, "Welcome to the game!"}}
	sendMessage(ws, msg)
	return p.Id, nil
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
	if !p.ActiveConnection {
		return false
	}
	fmt.Println(p)
	fmt.Printf("%s hand:\n", p.Username)
	defer func() {
		var hands []OtherHandStruct
		for _, player := range p.Game.Players {
			if player.ActiveConnection {
				hands = append(hands, OtherHandStruct{player.Id, player.Hand, time.Now()})
			}

		}
		msg := WebSocketMessage{MessageTypeOtherHands, hands}
		for _, player := range p.Game.Players {
			if player.Id != p.Id && player.ActiveConnection {
				sendMessage(player.conn, msg)
			}
		}
	}()
	for _, card := range p.Hand {
		fmt.Printf("%s, ", card)
	}
	fmt.Println()
	count := p.EvalHand()
	if count == 21 {
		// p.Budget += p.Stake * 2.5
		fmt.Printf("Count: %d\n-----------\n", p.Count)

		fmt.Println("You have BLACKJACK. Congrats...")
		message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: false, Count: count}}
		sendMessage(p.conn, message)
		return false
	} else if count > 21 {
		// p.Stake = 0
		fmt.Printf("Count: %d\n-----------\n", p.Count)
		fmt.Println("Your hand exceeded the 21 score. You lost...")
		message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: false, Count: count}}
		sendMessage(p.conn, message)
		return false
	}
	fmt.Printf("Count: %d\n-----------\n", p.Count)

	message := WebSocketMessage{MessageTypeSendHand, HandMessage{Hand: p.Hand, Stage: true, Count: count}}
	sendMessage(p.conn, message)

	return true
}

func (d *Dealer) ShowHand() {
	fmt.Printf("Dealer's hand:\n")
	for _, card := range d.Hand {
		fmt.Printf("%s, ", card)
		// if i == 0 || d.GameStage {
		// 	fmt.Printf("%s, ", card)
		// } else {
		// 	fmt.Printf("***HIDDEN***")
		// }
	}

	if d.Game.GameStage == PostGame {
		message := WebSocketMessage{MessageTypeDealerFinalHand, HandMessage{Count: d.EvalHand(), Hand: d.Hand, Stage: true}}
		for _, player := range d.Game.Players {
			if !player.ActiveConnection {
				continue
			}
			sendMessage(player.conn, message)
		}
		d.EvalHand()
		fmt.Println(d.Count, "\n-------")
	} else {
		message := WebSocketMessage{MessageTypeDealerInitHand, d.Hand}
		for _, player := range d.Game.Players {
			if !player.ActiveConnection {
				continue
			}
			sendMessage(player.conn, message)
		}
		fmt.Println("\n-------")
	}
}

func (p *Player) SendReconnectState() {

	if p.Game.GameStage == PreGame {

		msg := WebSocketMessage{MessageTypeReconnectState, ReconnectStateMessage{
			DealerHand: []deck.Card{},
			Hand:       []deck.Card{},
			OtherHands: []OtherHandStruct{},
			Turn:       false,
			GameStage:  PreGame,
			Count:      p.EvalHand(),
		}}

		sendMessage(p.conn, msg)
		return
	}
	var index int
	var turn bool
	for i, player := range p.Game.Players {
		if player.Id == p.Id {
			index = i
			break
		}
	}
	turn = index != p.Game.CurrentTurn

	var hands []OtherHandStruct
	for _, player := range p.Game.Players {
		if player.ActiveConnection {
			hands = append(hands, OtherHandStruct{player.Id, player.Hand, time.Now()})
		}
	}

	msg := WebSocketMessage{MessageTypeReconnectState, ReconnectStateMessage{
		DealerHand: p.Game.Dealer.Hand,
		Hand:       p.Hand,
		OtherHands: hands,
		Turn:       turn,
		GameStage:  p.Game.GameStage,
		Count:      p.EvalHand(),
	}}

	sendMessage(p.conn, msg)

}

// func (d *Dealer) ResetDealer() {

// }
